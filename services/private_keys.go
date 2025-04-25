// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2020 Cory Slep
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package services

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/allinbits/apcore/models"
	"github.com/allinbits/apcore/paths"
	"github.com/allinbits/apcore/util"
)

const (
	minKeySize = 1024
)

const (
	pKeyHttpSigPurpose = "http-signature"
)

type PrivateKeys struct {
	Scheme      string
	Host        string
	DB          *sql.DB
	PrivateKeys *models.PrivateKeys
}

func (p *PrivateKeys) GetUserHTTPSignatureKey(c util.Context, userID paths.UUID) (k *rsa.PrivateKey, iri *url.URL, err error) {
	var kb []byte
	err = doInTx(c, p.DB, func(tx *sql.Tx) error {
		kb, err = p.PrivateKeys.GetByUserID(c, tx, string(userID), pKeyHttpSigPurpose)
		return err
	})
	if err != nil {
		return
	}
	var pk crypto.PrivateKey
	pk, err = deserializeRSAPrivateKey(kb)
	var ok bool
	k, ok = pk.(*rsa.PrivateKey)
	if !ok {
		err = errors.New("private key is not of type *rsa.PrivateKey")
		return
	}
	iri = paths.UUIDIRIFor(p.Scheme, p.Host, paths.HttpSigPubKeyKey, userID)
	return
}

func (p *PrivateKeys) GetUserHTTPSignatureKeyForInstanceActor(c util.Context) (k *rsa.PrivateKey, iri *url.URL, err error) {
	var kb []byte
	err = doInTx(c, p.DB, func(tx *sql.Tx) error {
		kb, err = p.PrivateKeys.GetInstanceActor(c, tx, pKeyHttpSigPurpose)
		return err
	})
	if err != nil {
		return
	}
	var pk crypto.PrivateKey
	pk, err = deserializeRSAPrivateKey(kb)
	var ok bool
	k, ok = pk.(*rsa.PrivateKey)
	if !ok {
		err = errors.New("private key is not of type *rsa.PrivateKey")
		return
	}
	iri = paths.ActorIRIFor(p.Scheme, p.Host, paths.HttpSigPubKeyKey, paths.InstanceActor)
	return
}

// CreateKeyFile writes a symmetric key of random bytes to a file.
func CreateKeyFile(file string) (err error) {
	c := 32
	k := make([]byte, c)
	var n int
	n, err = rand.Read(k)
	if err != nil {
		return
	} else if n != c {
		err = fmt.Errorf("crypto/rand read %d of %d bytes", n, c)
		return
	}
	err = ioutil.WriteFile(file, k, os.FileMode(0660))
	return
}

// createandSerializeRSAKeys creates a new RSA Private key of a given size
// and returns its PKCS8 encoded form and the public key's PEM form.
func createAndSerializeRSAKeys(n int) (priv []byte, pub string, err error) {
	var k *rsa.PrivateKey
	k, err = createRSAPrivateKey(n)
	if err != nil {
		return
	}
	priv, err = serializeRSAPrivateKey(k)
	if err != nil {
		return
	}
	pub, err = marshalPublicKey(&(k.PublicKey))
	return
}

// createRSAPrivateKey creates a new RSA Private key of a given size.
//
// Returns an error if the size is less than minKeySize.
func createRSAPrivateKey(n int) (k *rsa.PrivateKey, err error) {
	if n < minKeySize {
		err = fmt.Errorf("Creating a key of size < %d is forbidden: %d", minKeySize, n)
		return
	}
	k, err = rsa.GenerateKey(rand.Reader, n)
	return
}

// marshalPublicKey encodes a public key into PEM format.
func marshalPublicKey(p crypto.PublicKey) (string, error) {
	pkix, err := x509.MarshalPKIXPublicKey(p)
	if err != nil {
		return "", err
	}
	pb := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkix,
	})
	return string(pb), nil
}

// serializeRSAPrivateKey encodes a private key into PKCS8 format.
func serializeRSAPrivateKey(k *rsa.PrivateKey) ([]byte, error) {
	return x509.MarshalPKCS8PrivateKey(k)
}

// deserializeRSAPrivateKey decodes a private key from PKCS8 format.
func deserializeRSAPrivateKey(b []byte) (crypto.PrivateKey, error) {
	return x509.ParsePKCS8PrivateKey(b)
}
