// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package engine

import (
	"fmt"

	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol"
)

type RevokeMode int

const (
	RevokeKey RevokeMode = iota
	RevokeDevice
)

type RevokeEngine struct {
	libkb.Contextified
	deviceID keybase1.DeviceID
	kid      keybase1.KID
	mode     RevokeMode
	force    bool
}

type RevokeDeviceEngineArgs struct {
	ID    keybase1.DeviceID
	Force bool
}

func NewRevokeDeviceEngine(args RevokeDeviceEngineArgs, g *libkb.GlobalContext) *RevokeEngine {
	return &RevokeEngine{
		deviceID:     args.ID,
		mode:         RevokeDevice,
		force:        args.Force,
		Contextified: libkb.NewContextified(g),
	}
}

func NewRevokeKeyEngine(kid keybase1.KID, g *libkb.GlobalContext) *RevokeEngine {
	return &RevokeEngine{
		kid:          kid,
		mode:         RevokeKey,
		Contextified: libkb.NewContextified(g),
	}
}

func (e *RevokeEngine) Name() string {
	return "Revoke"
}

func (e *RevokeEngine) Prereqs() Prereqs {
	return Prereqs{
		Device: true,
	}
}

func (e *RevokeEngine) RequiredUIs() []libkb.UIKind {
	return []libkb.UIKind{
		libkb.LogUIKind,
		libkb.SecretUIKind,
	}
}

func (e *RevokeEngine) SubConsumers() []libkb.UIConsumer {
	return []libkb.UIConsumer{}
}

func (e *RevokeEngine) getKIDsToRevoke(me *libkb.User) ([]keybase1.KID, error) {
	if e.mode == RevokeDevice {
		deviceKeys, err := me.GetComputedKeyFamily().GetAllActiveKeysForDevice(e.deviceID)
		if err != nil {
			return nil, err
		}
		return deviceKeys, nil
	} else if e.mode == RevokeKey {
		kid := e.kid
		key, err := me.GetComputedKeyFamily().FindKeyWithKIDUnsafe(kid)
		if err != nil {
			return nil, err
		}
		if !libkb.IsPGP(key) {
			return nil, fmt.Errorf("Key %s is not a PGP key. To revoke device keys, use the `device remove` command.", e.kid)
		}
		for _, activePGPKey := range me.GetComputedKeyFamily().GetActivePGPKeys(false /* sibkeys only */) {
			if activePGPKey.GetKID().Equal(kid) {
				return []keybase1.KID{kid}, nil
			}
		}
		return nil, fmt.Errorf("PGP key %s is not active", e.kid)
	} else {
		return nil, fmt.Errorf("Unknown revoke mode: %d", e.mode)
	}
}

func (e *RevokeEngine) Run(ctx *Context) error {
	currentDevice := e.G().Env.GetDeviceID()
	var deviceID keybase1.DeviceID
	if e.mode == RevokeDevice {
		deviceID = e.deviceID
		if e.deviceID == currentDevice && !e.force {
			return fmt.Errorf("Can't revoke the current device.")
		}
	}

	me, err := libkb.LoadMe(libkb.NewLoadUserArg(e.G()))
	if err != nil {
		return err
	}

	kidsToRevoke, err := e.getKIDsToRevoke(me)
	if err != nil {
		return err
	}
	ctx.LogUI.Info("Revoking KIDs:")
	for _, kid := range kidsToRevoke {
		ctx.LogUI.Info("  %s", kid)
	}

	sigKey, err := e.G().Keyrings.GetSecretKeyWithPrompt(ctx.LoginContext, libkb.SecretKeyArg{
		Me:      me,
		KeyType: libkb.DeviceSigningKeyType,
	}, ctx.SecretUI, "to revoke another key")
	if err != nil {
		return err
	}
	if err = sigKey.CheckSecretKey(); err != nil {
		return err
	}

	proof, err := me.RevokeKeysProof(sigKey, kidsToRevoke, deviceID)
	if err != nil {
		return err
	}
	sig, _, _, err := libkb.SignJSON(proof, sigKey)
	if err != nil {
		return err
	}
	kid := sigKey.GetKID()
	_, err = e.G().API.Post(libkb.APIArg{
		Endpoint:    "sig/revoke",
		NeedSession: true,
		Args: libkb.HTTPArgs{
			"signing_kid": libkb.S{Val: kid.String()},
			"sig":         libkb.S{Val: sig},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
