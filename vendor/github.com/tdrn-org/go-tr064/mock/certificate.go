/*
 * Copyright 2024 Holger de Carne
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mock

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"time"
)

var mockCertificate *x509.Certificate
var mockKey crypto.PrivateKey

func mockCertificatePEM() []byte {
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: mockCertificate.Raw,
	}
	return pem.EncodeToMemory(block)
}

func mockKeyPEM() []byte {
	encodedKey, err := x509.MarshalPKCS8PrivateKey(mockKey)
	if err != nil {
		panic(err)
	}
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: encodedKey,
	}
	return pem.EncodeToMemory(block)
}

// Generate self-signed certificate.
func init() {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "mock"},
		NotBefore:    now,
		NotAfter:     now.AddDate(0, 0, 1),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		IsCA:         true,
	}
	certificateBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		log.Fatal(err)
	}
	certificate, err := x509.ParseCertificate(certificateBytes)
	if err != nil {
		log.Fatal(err)
	}
	mockCertificate = certificate
	mockKey = key
}
