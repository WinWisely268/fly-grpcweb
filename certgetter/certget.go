package main

import (
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
)

const defaultOutputPath = "./cacert.pem"

var (
	outputPath = ""
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "caget <URL>",
		Short: "caget <URL>",
		Args:  cobra.ExactArgs(1),
	}
	rootCmd.PersistentFlags().StringVarP(&outputPath, "output", "o", defaultOutputPath, "output path for the CA cert")

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		conf := &tls.Config{
			InsecureSkipVerify: true,
		}
		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", args[0]), conf)
		if err != nil {
			log.Println("Error while dialing", err)
			return err
		}
		defer conn.Close()
		certs := conn.ConnectionState().PeerCertificates
		for _, cert := range certs {
			if cert.IsCA {
				publicKeyBlock := pem.Block{
					Type:  "CERTIFICATE",
					Bytes: cert.Raw,
				}
				publicKeyPem := pem.EncodeToMemory(&publicKeyBlock)
				err = ioutil.WriteFile(outputPath, publicKeyPem, 0644)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
