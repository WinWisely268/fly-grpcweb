package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	client "github.com/getcouragenow/protoc-gen-cobra/client"
	"github.com/getcouragenow/protoc-gen-cobra/iocodec"
	"github.com/getcouragenow/protoc-gen-cobra/naming"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"grpc-test/rpc"
	"io/ioutil"
	"log"
	"time"
)

const (
	defaultTimeout = 5 * time.Second
)

func main() {
	srvCfg := client.WithServerAddr("fly-.fly.dev:443")
	encoder := iocodec.JSONEncoderMaker(true)
	jsonCfg := client.WithOutputEncoder("json", encoder)
	flagBinder := func(fs *pflag.FlagSet, namer naming.Namer) {
		fs.StringVarP(&MainProxyCLIConfig.AccessKey, namer("JWT Access Token"), "j", MainProxyCLIConfig.AccessKey, "JWT Access Token")
	}
	preDialer := func(_ context.Context, opts *[]grpc.DialOption) error {
		cfg := MainProxyCLIConfig
		tkn := &oauth2.Token{
			TokenType: "Bearer",
		}
		if cfg.AccessKey != "" {
			tkn.AccessToken = cfg.AccessKey
			cred := oauth.NewOauthAccess(tkn)
			*opts = append(*opts, grpc.WithPerRPCCredentials(cred))
		}
		return nil
	}
	client.RegisterFlagBinder(flagBinder)
	client.RegisterPreDialer(preDialer)
	// creds, err := getRemoteCACert("sparkling-snow-2014.fly.dev")
	// if err != nil {
	// 	log.Fatal(fmt.Errorf("unable to load CA Root path: %v", err))
	// }
	caCfg := client.WithTLSCACertFile("certs/rootca.pem")
	clientOptions := []client.Option{
		srvCfg, jsonCfg, caCfg,
	}
	helloClient := rpc.MainServiceClientCommand(clientOptions...)
	if err := helloClient.Execute(); err != nil {
		log.Fatal(err)
	}
}

func clientLoadCA(cacertPath string) (credentials.TransportCredentials, error) {
	pemServerCA, err := ioutil.ReadFile(cacertPath)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}
	config := &tls.Config{
		RootCAs: certPool,
	}
	return credentials.NewTLS(config), nil
}

func getRemoteCACert(domain string) (credentials.TransportCredentials, error) {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", domain), conf)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	for _, cert := range certs {
		if cert.IsCA {
			certPool := x509.NewCertPool()
			certPool.AddCert(cert)
			config := &tls.Config{
				RootCAs: certPool,
			}
			return credentials.NewTLS(config), nil
		}
	}
	return nil, fmt.Errorf("unable to get CA from server: cert does not exist")
}

type mainProxyCliConfig struct {
	AccessKey string
}

var MainProxyCLIConfig = &mainProxyCliConfig{}

