// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package saltpack

import (
	"bytes"
	"io"
)

type testSignOptions struct {
	corruptHeader func(sh *SignatureHeader)
	swapBlock     bool
	skipBlock     func(blockNum PacketSeqno) bool
	skipFooter    bool
}

type testSignStream struct {
	header      *SignatureHeader
	wroteHeader bool
	encoder     encoder
	buffer      bytes.Buffer
	block       []byte
	seqno       PacketSeqno
	secretKey   SigningSecretKey
	options     testSignOptions
	savedBlock  *SignatureBlock
}

func newTestSignStream(w io.Writer, signer SigningSecretKey, opts testSignOptions) (*testSignStream, error) {
	if signer == nil {
		return nil, ErrInvalidParameter{message: "no signing key provided"}
	}

	header, err := newSignatureHeader(signer.PublicKey(), MessageTypeAttachedSignature)
	if err != nil {
		return nil, err
	}
	if opts.corruptHeader != nil {
		opts.corruptHeader(header)
	}

	stream := &testSignStream{
		header:    header,
		encoder:   newEncoder(w),
		block:     make([]byte, SignatureBlockSize),
		secretKey: signer,
		options:   opts,
	}

	return stream, nil
}

func (s *testSignStream) Write(p []byte) (int, error) {
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

func (s *testSignStream) Close() error {
	for s.buffer.Len() > 0 {
		if err := s.signBlock(); err != nil {
			return err
		}
	}

	if s.options.skipFooter {
		return nil
	}

	return s.writeFooter()
}

func (s *testSignStream) signBlock() error {
	n, err := s.buffer.Read(s.block[:])
	if err != nil {
		return err
	}
	return s.signBytes(s.block[:n])
}

func (s *testSignStream) signBytes(b []byte) error {
	block := SignatureBlock{
		PayloadChunk: b,
		seqno:        s.seqno,
	}
	sig, err := s.computeSig(&block)
	if err != nil {
		return err
	}
	block.Signature = sig

	if s.options.swapBlock {
		if s.seqno == 0 {
			s.savedBlock = &block
			s.seqno++
			return nil
		}
	}

	if s.options.skipBlock == nil || !s.options.skipBlock(s.seqno) {
		if err := s.encoder.Encode(block); err != nil {
			return err
		}
		s.seqno++
	}

	if s.options.swapBlock {
		if s.savedBlock != nil {
			if err := s.encoder.Encode(*s.savedBlock); err != nil {
				return err
			}
			s.savedBlock = nil
			return nil
		}
	}

	return nil
}

func (s *testSignStream) writeFooter() error {
	return s.signBytes([]byte{})
}

func (s *testSignStream) computeSig(block *SignatureBlock) ([]byte, error) {
	return s.secretKey.Sign(computeAttachedDigest(s.header.Nonce, block))
}

func testTweakSign(plaintext []byte, signer SigningSecretKey, opts testSignOptions) ([]byte, error) {
	var buf bytes.Buffer
	s, err := newTestSignStream(&buf, signer, opts)
	if err != nil {
		return nil, err
	}
	if _, err := s.Write(plaintext); err != nil {
		return nil, err
	}
	if err := s.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func testTweakSignDetached(plaintext []byte, signer SigningSecretKey, opts testSignOptions) ([]byte, error) {
	if signer == nil {
		return nil, ErrInvalidParameter{message: "no signing key provided"}
	}
	header, err := newSignatureHeader(signer.PublicKey(), MessageTypeDetachedSignature)
	if err != nil {
		return nil, err
	}

	signature, err := signer.Sign(computeDetachedDigest(header.Nonce, plaintext))
	if err != nil {
		return nil, err
	}
	header.Signature = signature

	if opts.corruptHeader != nil {
		opts.corruptHeader(header)
	}

	return encodeToBytes(header)
}
