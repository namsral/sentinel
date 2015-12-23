// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apn

import (
	"bytes"
	"container/list"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

var (
	APNHost     = "gateway.sandbox.push.apple.com"
	APNPort     = "2195"
	APNPassword string
	TLSCAPath   string
)

func init() {
	if s := os.Getenv("APN_HOST"); s != "" {
		APNHost = s
	}
	if s := os.Getenv("APN_PORT"); s != "" {
		APNPort = s
	}
	APNPassword = os.Getenv("APN_PASSWORD")
	TLSCAPath = os.Getenv("TLSCA_PATH")
}

func X509KeyPair(certPEMBlock, keyPEMBlock, pw []byte) (cert tls.Certificate, err error) {
	var certDERBlock *pem.Block
	for {
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		}
	}

	if len(cert.Certificate) == 0 {
		err = errors.New("crypto/tls: failed to parse certificate PEM data")
		return
	}
	var keyDERBlock *pem.Block
	for {
		keyDERBlock, keyPEMBlock = pem.Decode(keyPEMBlock)
		if keyDERBlock == nil {
			err = errors.New("crypto/tls: failed to parse key PEM data")
			return
		}
		if x509.IsEncryptedPEMBlock(keyDERBlock) {
			out, err2 := x509.DecryptPEMBlock(keyDERBlock, pw)
			if err2 != nil {
				err = err2
				return
			}
			keyDERBlock.Bytes = out
			break
		}
		if keyDERBlock.Type == "PRIVATE KEY" || strings.HasSuffix(keyDERBlock.Type, " PRIVATE KEY") {
			break
		}
	}

	cert.PrivateKey, err = parsePrivateKey(keyDERBlock.Bytes)
	if err != nil {
		return
	}
	// We don't need to parse the public key for TLS, but we so do anyway
	// to check that it looks sane and matches the private key.
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return
	}

	switch pub := x509Cert.PublicKey.(type) {
	case *rsa.PublicKey:
		priv, ok := cert.PrivateKey.(*rsa.PrivateKey)
		if !ok {
			err = errors.New("crypto/tls: private key type does not match public key type")
			return
		}
		if pub.N.Cmp(priv.N) != 0 {
			err = errors.New("crypto/tls: private key does not match public key")
			return
		}
	case *ecdsa.PublicKey:
		priv, ok := cert.PrivateKey.(*ecdsa.PrivateKey)
		if !ok {
			err = errors.New("crypto/tls: private key type does not match public key type")
			return

		}
		if pub.X.Cmp(priv.X) != 0 || pub.Y.Cmp(priv.Y) != 0 {
			err = errors.New("crypto/tls: private key does not match public key")
			return
		}
	default:
		err = errors.New("crypto/tls: unknown public key algorithm")
		return
	}
	return
}

// Attempt to parse the given private key DER block. OpenSSL 0.9.8 generates
// PKCS#1 private keys by default, while OpenSSL 1.0.0 generates PKCS#8 keys.
// OpenSSL ecparam generates SEC1 EC private keys for ECDSA. We try all three.
func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, errors.New("crypto/tls: found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	return nil, errors.New("crypto/tls: failed to parse private key")
}

func createSocket() (net.Conn, error) {
	certBytes, err := ioutil.ReadFile(TLSCAPath)
	if err != nil {
		return nil, err
	}

	keyBytes, err := ioutil.ReadFile(TLSCAPath)
	if err != nil {
		return nil, err
	}

	//password := "1234567"
	/*
	   block, rest := pem.Decode(keyBytes)
	   if len(rest) > 0 {
	       // return nil, errors.New("extra data")
	   }

	   if block == nil {
	       return nil, errors.New("NO PEM FOUND")
	   }
	   fmt.Printf("%#v\n", block)

	   der, err := x509.DecryptPEMBlock(block, []byte(password))
	   if err != nil {
	       return nil, err

	   }

	   fmt.Println("x509:1")*/

	x509Cert, err := X509KeyPair(certBytes, keyBytes, []byte(APNPassword))
	if err != nil {
		return nil, err
	}

	// serverName := "gateway.sandbox.push.apple.com"
	// serverName := "gateway.push.apple.com"

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{x509Cert},
		ServerName:   APNHost,
	}

	fmt.Println("x509:2")
	tcpSocket, err := net.Dial("tcp", net.JoinHostPort(APNHost, APNPort))
	if err != nil {
		//failed to connect to gateway
		panic(err)
	}

	tlsSocket := tls.Client(tcpSocket, tlsConf)
	err = tlsSocket.Handshake()
	if err != nil {
		//failed to handshake with tls information
		panic(err)
	}
	fmt.Println("x509:3")

	return tlsSocket, nil
}

type PushNotification struct {
	AlertText string
	Token     string
	id        uint32
}

func (pn *PushNotification) toBytes() []byte {
	buffer := new(bytes.Buffer)
	frameBuffer := new(bytes.Buffer)
	token, err := hex.DecodeString(pn.Token)
	if err != nil {
		//Failed to decode token
		panic(err)
	}
	payloadBytes := []byte("{\"aps\":{\"alert\":" + pn.AlertText + "}}")

	//write token
	binary.Write(buffer, binary.BigEndian, uint8(1))
	binary.Write(buffer, binary.BigEndian, uint16(32))
	binary.Write(buffer, binary.BigEndian, token)

	//write payload
	binary.Write(buffer, binary.BigEndian, uint8(2))
	binary.Write(buffer, binary.BigEndian, uint16(len(payloadBytes)))
	binary.Write(buffer, binary.BigEndian, payloadBytes)

	//write push notification id
	binary.Write(buffer, binary.BigEndian, uint8(3))
	binary.Write(buffer, binary.BigEndian, uint16(4))
	binary.Write(buffer, binary.BigEndian, pn.id)

	//write header info and item info for frame
	binary.Write(frameBuffer, binary.BigEndian, uint8(2))
	binary.Write(frameBuffer, binary.BigEndian, uint32(buffer.Len()))

	return frameBuffer.Bytes()
}

func socketReader(errChan chan *SocketClosed, socket net.Conn) {
	buffer := make([]byte, 6, 6)
	_, err := socket.Read(buffer)
	if err != nil {
		//the socket was closed but nothing was read
		errChan <- &SocketClosed{
			ErrorCode: 10,
			MessageID: 0,
		}
	} else {
		//apple sent us a response
		messageId := binary.BigEndian.Uint32(buffer[2:])
		errChan <- &SocketClosed{
			ErrorCode: uint8(buffer[1]),
			MessageID: messageId,
		}
	}
}

type WriterClosed struct {
	//what caused the writer to close
	SocketClosedObj *SocketClosed
	//any unsent notifications
	UnsentNotifications *list.List
}

func socketWriter(sendChan chan *PushNotification, errChan chan *SocketClosed,
	writerClosedChan chan *WriterClosed, socket net.Conn) {
	inFlightNotifications := list.New()
	nextId := uint32(1)
	shouldClose := false
	var socketClosed *SocketClosed
	for {
		if shouldClose {
			break
		}
		select {
		case pn := <-sendChan:
			pn.id = nextId

			inFlightNotifications.PushFront(pn)
			//check to see if we've overrun our buffer
			//if so, remove one from the buffer
			if inFlightNotifications.Len() > 1000 {
				inFlightNotifications.Remove(inFlightNotifications.Back())
			}

			fmt.Println("Writing to socket")

			socket.Write(pn.toBytes())
			nextId++
			break
		case socketClosed = <-errChan:
			shouldClose = true
			break
		}
	}

	unsentNotifications := list.New()
	if socketClosed.MessageID > 0 {
		//we received error
		for e := inFlightNotifications.Front(); e != nil; e = e.Next() {
			pn := e.Value.(*PushNotification)
			if pn.id == socketClosed.MessageID {
				break
			}
			unsentNotifications.PushFront(pn)
		}
	}

	writerClosedChan <- &WriterClosed{
		SocketClosedObj:     socketClosed,
		UnsentNotifications: unsentNotifications,
	}
}

type SocketClosed struct {
	//Internal ID of the message that caused the error
	MessageID uint32
	//Error code returned by Apple
	ErrorCode uint8
}

func SendMessage() error {
	socket, err := createSocket()
	if err != nil {
		return err
	}

	sendChan := make(chan *PushNotification)
	errChan := make(chan *SocketClosed)
	writerClosedChan := make(chan *WriterClosed)

	fmt.Println("Pushing message")

	go socketReader(errChan, socket)
	go socketWriter(sendChan, errChan, writerClosedChan, socket)

	sendChan <- &PushNotification{AlertText: "Test", Token: "e2f28cb73f1f47dd171f6749857ded7c462af8d54193893e34ec0ed51dfcee25"}

	fmt.Println("waiting message")
	select {
	case err := <-errChan:
		fmt.Printf("errch%#v", err)
	case wc := <-writerClosedChan:
		fmt.Printf("wc %#v", wc)

	}
	fmt.Println("waited message")
	return nil
}
