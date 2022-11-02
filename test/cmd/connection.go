/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"crypto/x509"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	mspID         = "Org1MSP"
	clientCertEnv = "CLIENT_CERT"
	clientKeyEnv  = "CLIENT_KEY"
	caCertEnv     = "CA_CERT"
	endpointEnv   = "ENDPOINT"
)

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	// certificate, err := loadCertificate(os.Getenv(caCertEnv))
	// if err != nil {
	// 	panic(err)
	// }

	// certPool := x509.NewCertPool()
	// certPool.AddCert(certificate)
	// transportCredentials := credentials.NewClientTLSFromCert(certPool, "")

	transportCredentials := insecure.NewCredentials()

	connection, err := grpc.Dial(os.Getenv(endpointEnv), grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificate, err := loadCertificate(os.Getenv(clientCertEnv))
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename) //#nosec G304 -- test input
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	privateKeyPEM, err := os.ReadFile(os.Getenv(clientKeyEnv))

	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}
