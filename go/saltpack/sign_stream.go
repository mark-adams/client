// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package saltpack

import (
	"bytes"
	"crypto/sha512"
	"hash"
	"io"
)

type signAttachedStream struct {
	header      *SignatureHeader
	wroteHeader bool
	encoder     encoder
	buffer      bytes.Buffer
	block       []byte
	seqno       PacketSeqno
	secretKey   SigningSecretKey
}

func newSignAttachedStream(w io.Writer, signer SigningSecretKey) (*signAttachedStream, error) {
	if signer == nil {
		return nil, ErrInvalidParameter{message: "no signing key provided"}
	}

	header, err := newSignatureHeader(signer.PublicKey(), MessageTypeAttachedSignature)
	if err != nil {
		return nil, err
	}

	stream := &signAttachedStream{
		header:    header,
		encoder:   newEncoder(w),
		block:     make([]byte, SignatureBlockSize),
		secretKey: signer,
	}

	return stream, nil
}

func (s *signAttachedStream) Write(p []byte) (int, error) {
	if !s.wroteHeader {
		s.wroteHeader = true
		if err := s.encoder.Encode(s.header); err != nil {
			return 0, err
		}
	}

	n, err := s.buffer.Write(p)
	if err != nil {
		return 0, err
	}

	for s.buffer.Len() >= SignatureBlockSize {
		if err := s.signBlock(); err != nil {
			return 0, err
		}
	}

	return n, nil
}

func (s *signAttachedStream) Close() error {
	if !s.wroteHeader {
		s.wroteHeader = true
		if err := s.encoder.Encode(s.header); err != nil {
			return err
		}
	}

	for s.buffer.Len() > 0 {
		if err := s.signBlock(); err != nil {
			return err
		}
	}
	return s.writeFooter()
}

func (s *signAttachedStream) signBlock() error {
	n, err := s.buffer.Read(s.block[:])
	if err != nil {
		return err
	}
	return s.signBytes(s.block[:n])
}

func (s *signAttachedStream) signBytes(b []byte) error {
	block := SignatureBlock{
		PayloadChunk: b,
		seqno:        s.seqno,
	}
	sig, err := s.computeSig(&block)
	if err != nil {
		return err
	}
	block.Signature = sig

	if err := s.encoder.Encode(block); err != nil {
		return err
	}

	s.seqno++
	return nil
}

func (s *signAttachedStream) writeFooter() error {
	return s.signBytes([]byte{})
}

func (s *signAttachedStream) computeSig(block *SignatureBlock) ([]byte, error) {
	return s.secretKey.Sign(computeAttachedDigest(s.header.Nonce, block))
}

type signDetachedStream struct {
	header    *SignatureHeader
	encoder   encoder
	secretKey SigningSecretKey
	hasher    hash.Hash
}

func newSignDetachedStream(w io.Writer, signer SigningSecretKey) (*signDetachedStream, error) {
	if signer == nil {
		return nil, ErrInvalidParameter{message: "no signing key provided"}
	}

	header, err := newSignatureHeader(signer.PublicKey(), MessageTypeDetachedSignature)
	if err != nil {
		return nil, err
	}

	stream := &signDetachedStream{
		header:    header,
		encoder:   newEncoder(w),
		secretKey: signer,
		hasher:    sha512.New(),
	}

	stream.hasher.Write(stream.header.Nonce)

	return stream, nil
}

func (s *signDetachedStream) Write(p []byte) (int, error) {
	return s.hasher.Write(p)
}

func (s *signDetachedStream) Close() error {
	signature, err := s.secretKey.Sign(detachedDigest(s.hasher.Sum(nil)))
	if err != nil {
		return err
	}
	s.header.Signature = signature

	if err := s.encoder.Encode(s.header); err != nil {
		return err
	}

	return nil
}
