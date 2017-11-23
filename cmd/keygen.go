/*
 * Copyright (c) 2017.  Harrison Zhu <wcg6121@gmail.com>
 * This file is part of djob <https://github.com/HZ89/djob>.
 *
 * djob is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * djob is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with djob.  If not, see <http://www.gnu.org/licenses/>.
 */

package cmd

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/cli"
)

// key command type, implementation Command interface
type KeygenCmd struct {
	Ui     cli.Ui
	config *config
}

type config struct {
	host     string
	validFor time.Duration
	initCa   bool
	ca       string
	caKey    string
}

// prepare config
func newConfig(args []string) *config {
	flags := flag.NewFlagSet("generate", flag.ContinueOnError)
	var (
		host     = flags.String("host", "", "Comma-separated hostnames and IPs to generate a certificate for")
		validFor = flags.Duration("duration", 365*24, "Duration that certificate is valid for, unit is hour")
		initCa   = flags.Bool("initca", false, "Create a root ca keypair")
		ca       = flags.String("ca", "./ca.pem", "Ca public key")
		caKey    = flags.String("cakey", "./ca-key.pem", "Ca private key")
	)
	flags.Parse(args)
	return &config{
		host:     *host,
		validFor: *validFor,
		initCa:   *initCa,
		ca:       *ca,
		caKey:    *caKey,
	}
}

// generate tls certificate, if have no ca certificate it can generate new ca.
func (c *KeygenCmd) genCert() int {
	if len(c.config.host) == 0 {
		c.Ui.Error("Missing required --host parameter")
		return 1
	}

	notBefore := time.Now()

	var caPrivInPem []byte
	var caInPem []byte
	var err error

	if c.config.initCa {

		notAfter := notBefore.Add(c.config.validFor * 10 * time.Hour)
		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("failed to generate serial number: %s", err.Error()))
			return 1
		}

		caTemplate := x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				Organization: []string{"djob ca"},
			},
			NotBefore: notBefore,
			NotAfter:  notAfter,

			KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,

			IsCA: true,
		}
		caPrivInPem, caInPem, err = c.createKeypair(true, &caTemplate, nil, nil)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		err = ioutil.WriteFile("ca.pem", caInPem, 0644)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to writing ca.pem: %s", err.Error()))
		}
		c.Ui.Output("Write root certificate ca.pem succeed")
		err = ioutil.WriteFile("ca-key.pem", caPrivInPem, 0600)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to writing ca-key.pem: %s", err.Error()))
		}
		c.Ui.Output("Write root key ca-key.pem succeed")

	} else {
		if len(c.config.ca) == 0 || len(c.config.caKey) == 0 {
			c.Ui.Error(fmt.Sprintf("Missing required --ca or --cakey parameter"))
			return 1
		}
		caInPem, err = ioutil.ReadFile(c.config.ca)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to open ca file %s: %s", c.config.ca, err.Error()))
			return 1
		}
		caPrivInPem, err = ioutil.ReadFile(c.config.caKey)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to open ca file %s: %s", c.config.caKey, err.Error()))
			return 1
		}
	}

	notAfter := notBefore.Add(c.config.validFor * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("failed to generate serial number: %s", err.Error()))
		return 1
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	hosts := strings.Split(c.config.host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	pkInPem, certInPem, err := c.createKeypair(false, &template, caInPem, caPrivInPem)
	if err != nil {
		c.Ui.Error(fmt.Sprintf(err.Error()))
		return 1
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s.pem", hosts[0]), certInPem, 0644)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to writing file %s.pem: %s", hosts[0], err.Error()))
		return 1
	}
	c.Ui.Output(fmt.Sprintf("Write certificate %s.pem succeed", hosts[0]))
	err = ioutil.WriteFile(fmt.Sprintf("%s-key.pem", hosts[0]), pkInPem, 0600)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to writing file %s.pem: %s", hosts[0], err.Error()))
	}
	c.Ui.Output(fmt.Sprintf("Write key file %s-key.pem succeed", hosts[0]))
	return 0
}

func (c *KeygenCmd) createKeypair(selfSign bool, template *x509.Certificate, parentPem []byte, privPem []byte) (pkInPem []byte, certInPem []byte, err error) {
	var priv *ecdsa.PrivateKey
	var parent *x509.Certificate
	newPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return pkInPem, certInPem, err
	}
	newPub := publicKey(newPriv)
	if selfSign {
		parent = template
		priv = newPriv
	} else {
		privBlock, _ := pem.Decode(privPem)
		if privBlock == nil {
			return pkInPem, certInPem, errors.New("failed to decode private pem")
		}
		priv, err = x509.ParseECPrivateKey(privBlock.Bytes)
		if err != nil {
			return pkInPem, certInPem, err
		}
		parentBlock, _ := pem.Decode(parentPem)
		if privBlock == nil {
			return pkInPem, certInPem, errors.New("failed to decode certificate pem")
		}
		cert, err := x509.ParseCertificate(parentBlock.Bytes)
		if err != nil {
			return pkInPem, certInPem, err
		}
		p := x509.Certificate{
			SerialNumber:          cert.SerialNumber,
			Subject:               cert.Subject,
			NotBefore:             cert.NotBefore,
			NotAfter:              cert.NotAfter,
			KeyUsage:              cert.KeyUsage,
			ExtKeyUsage:           cert.ExtKeyUsage,
			BasicConstraintsValid: cert.BasicConstraintsValid,
			IsCA: cert.IsCA,
		}
		parent = &p
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, template, parent, newPub, priv)
	if err != nil {
		return pkInPem, certInPem, err
	}
	var certbuf, pkbuf bytes.Buffer
	pem.Encode(&certbuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certInPem = certbuf.Bytes()
	pem.Encode(&pkbuf, pemBlockForKey(newPriv))
	pkInPem = pkbuf.Bytes()
	return
}

// generate serf encryption key
func (c *KeygenCmd) genKey() int {
	key := make([]byte, 16)
	n, err := rand.Reader.Read(key)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed reading random data: %s", err.Error()))
		return 1
	}
	if n != 16 {
		c.Ui.Error(fmt.Sprintf("Failed read enough entropy. try agine"))
		return 1
	}
	c.Ui.Output(base64.StdEncoding.EncodeToString(key))
	return 0
}

// return help info
func (c *KeygenCmd) Synopsis() string {
	return "Generates tls files or encryption key"
}

// custom help info
func (c *KeygenCmd) Help() string {
	helpText := `
	Usage: djob generate cert [options]
	       djob generate key
	  Generates a new encryption that can be used to configure the agent to encrypt traffic.
	  The output of key command is already in the proper format that the agent expects.
	  cert command will generates tls files. Include ca pk, pub key.
	options:
	  --help      This help
	  --host      Comma-separated hostnames and IPs to generate a certificate for
	  --duration  Duration that certificate is valid for, unit is hour
	  --initca    Create a root ca keypair
	  --ca        Ca public key
	  --cakey     Ca private key
	`
	return strings.TrimSpace(helpText)
}

// exec this command
func (c *KeygenCmd) Run(args []string) int {
	if len(args) > 0 {
		if args[0] == "cert" {
			c.config = newConfig(args[1:])
			return c.genCert()
		}
		if args[0] == "key" {
			return c.genKey()
		}
	}

	c.Ui.Error("Args Error")
	c.Ui.Output(c.Help())
	return 1
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}
