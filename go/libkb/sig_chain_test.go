// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package libkb

import (
	// "fmt"
	"encoding/json"
	"reflect"
	"sort"
	"testing"
	"time"

	keybase1 "github.com/keybase/client/go/protocol"
	jsonw "github.com/keybase/go-jsonw"
	testvectors "github.com/keybase/keybase-test-vectors/go"
)

// Returns a map from error name strings to sets of Go error types. If a test
// returns any error type in the corresponding set, it's a pass. (The reason
// the types aren't one-to-one here is that implementation differences between
// the Go and JS sigchains make that more trouble than it's worth.)
func getErrorTypesMap() map[string]map[reflect.Type]bool {
	return map[string]map[reflect.Type]bool{
		"CTIME_MISMATCH": {
			reflect.TypeOf(CtimeMismatchError{}): true,
		},
		"EXPIRED_SIBKEY": {
			reflect.TypeOf(KeyExpiredError{}): true,
		},
		"FINGERPRINT_MISMATCH": {
			reflect.TypeOf(ChainLinkFingerprintMismatchError{}): true,
		},
		"INVALID_SIBKEY": {
			reflect.TypeOf(KeyRevokedError{}): true,
		},
		"NO_KEY_WITH_THIS_HASH": {
			reflect.TypeOf(NoKeyError{}): true,
		},
		"KEY_OWNERSHIP": {
			reflect.TypeOf(KeyFamilyError{}): true,
		},
		"KID_MISMATCH": {
			reflect.TypeOf(ChainLinkKIDMismatchError{}): true,
		},
		"NONEXISTENT_KID": {
			reflect.TypeOf(KeyFamilyError{}): true,
		},
		"NOT_LATEST_SUBCHAIN": {
			reflect.TypeOf(NotLatestSubchainError{}): true,
		},
		"REVERSE_SIG_VERIFY_FAILED": {
			reflect.TypeOf(ReverseSigError{}): true,
		},
		"VERIFY_FAILED": {
			reflect.TypeOf(BadSigError{}): true,
		},
		"WRONG_UID": {
			reflect.TypeOf(UIDMismatchError{}): true,
		},
		"WRONG_USERNAME": {
			reflect.TypeOf(BadUsernameError{}): true,
		},
		"WRONG_SEQNO": {
			reflect.TypeOf(ChainLinkWrongSeqnoError{}): true,
		},
		"WRONG_PREV": {
			reflect.TypeOf(ChainLinkPrevHashMismatchError{}): true,
		},
	}
}

// One of the test cases from the JSON list of all tests.
type TestCase struct {
	Input   string `json:"input"`
	Len     int    `json:"len"`
	Sibkeys int    `json:"sibkeys"`
	Subkeys int    `json:"subkeys"`
	ErrType string `json:"err_type"`
	Eldest  string `json:"eldest"`
}

// The JSON list of all test cases.
type TestList struct {
	Tests      map[string]TestCase `json:"tests"`
	ErrorTypes []string            `json:"error_types"`
}

// The input data for a single test. Each tests has its own input JSON file.
type TestInput struct {
	// We omit the "chain" member here, because we need it in blob form.
	Username  string            `json:"username"`
	UID       string            `json:"uid"`
	Keys      []string          `json:"keys"`
	LabelKids map[string]string `json:"label_kids"`
	LabelSigs map[string]string `json:"label_sigs"`
}

func TestAllChains(t *testing.T) {
	tc := SetupTest(t, "test_all_chains")
	defer tc.Cleanup()

	var testList TestList
	json.Unmarshal([]byte(testvectors.ChainTests), &testList)
	// Always do the tests in alphabetical order.
	testNames := []string{}
	for name := range testList.Tests {
		testNames = append(testNames, name)
	}
	sort.Strings(testNames)
	for _, name := range testNames {
		testCase := testList.Tests[name]
		G.Log.Info("starting sigchain test case %s (%s)", name, testCase.Input)
		doChainTest(t, testCase)
	}
}

func doChainTest(t *testing.T, testCase TestCase) {
	inputJSON, exists := testvectors.ChainTestInputs[testCase.Input]
	if !exists {
		t.Fatal("missing test input: " + testCase.Input)
	}
	// Unmarshal test input in two ways: once for the structured data and once
	// for the chain link blobs.
	var input TestInput
	err := json.Unmarshal([]byte(inputJSON), &input)
	if err != nil {
		t.Fatal(err)
	}
	inputBlob, err := jsonw.Unmarshal([]byte(inputJSON))
	if err != nil {
		t.Fatal(err)
	}
	uid, err := UIDFromHex(input.UID)
	if err != nil {
		t.Fatal(err)
	}
	chainLen, err := inputBlob.AtKey("chain").Len()
	if err != nil {
		t.Fatal(err)
	}

	// Get the eldest key. This is assumed to be the first key in the list of
	// bundles, unless the "eldest" field is given in the test description, in
	// which case the eldest key is specified by name.
	var eldestKID keybase1.KID
	if testCase.Eldest == "" {
		eldestKey, err := ParseGenericKey(input.Keys[0])
		if err != nil {
			t.Fatal(err)
		}
		eldestKID = eldestKey.GetKID()
	} else {
		eldestKIDStr, found := input.LabelKids[testCase.Eldest]
		if !found {
			t.Fatalf("No KID found for label %s", testCase.Eldest)
		}
		eldestKID = keybase1.KIDFromString(eldestKIDStr)
	}

	// Parse all the key bundles.
	keyFamily, err := createKeyFamily(input.Keys)
	if err != nil {
		t.Fatal(err)
	}

	// Run the actual sigchain parsing and verification. This is most of the
	// code that's actually being tested.
	var sigchainErr error
	ckf := ComputedKeyFamily{kf: keyFamily}
	sigchain := SigChain{username: NewNormalizedUsername(input.Username), uid: uid, loadedFromLinkOne: true}
	for i := 0; i < chainLen; i++ {
		linkBlob := inputBlob.AtKey("chain").AtIndex(i)
		link, err := ImportLinkFromServer(&sigchain, linkBlob, uid)
		if err != nil {
			sigchainErr = err
			break
		}
		sigchain.chainLinks = append(sigchain.chainLinks, link)
	}
	if sigchainErr == nil {
		_, sigchainErr = sigchain.VerifySigsAndComputeKeys(eldestKID, &ckf)
	}

	// Some tests expect an error. If we get one, make sure it's the right
	// type.
	if testCase.ErrType != "" {
		if sigchainErr == nil {
			t.Fatalf("Expected %s error from VerifySigsAndComputeKeys. No error returned.", testCase.ErrType)
		}
		foundType := reflect.TypeOf(sigchainErr)
		expectedTypes := getErrorTypesMap()[testCase.ErrType]
		if expectedTypes == nil || len(expectedTypes) == 0 {
			msg := "No Go error types defined for expected failure %s.\n" +
				"This could be because of new test cases in github.com/keybase/keybase-test-vectors.\n" +
				"Go error returned: %s"
			t.Fatalf(msg, testCase.ErrType, foundType)
		}
		if expectedTypes[foundType] {
			// Success! We found the error we expected. This test is done.
			G.Log.Debug("EXPECTED error encountered", sigchainErr)
			return
		}

		// Got an error, but one of the wrong type. Tests with error names
		// that are missing from the map (maybe because we add new test
		// cases in the future) will also hit this branch.
		t.Fatalf("Wrong error type encountered. Expected %v (%s), got %s: %s",
			expectedTypes, testCase.ErrType, foundType, sigchainErr)

	}

	// Tests that expected an error terminated above. Tests that get here
	// should succeed without errors.
	if sigchainErr != nil {
		t.Fatal(err)
	}

	// Check the expected results: total unrevoked links, sibkeys, and subkeys.
	unrevokedCount := 0

	// XXX we should really contextify this
	idtable, err := NewIdentityTable(nil, eldestKID, &sigchain, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, link := range idtable.links {
		if !link.IsRevoked() {
			unrevokedCount++
		}
	}
	if unrevokedCount != testCase.Len {
		t.Fatalf("Expected %d unrevoked links, but found %d.", testCase.Len, unrevokedCount)
	}
	// Don't use the current time to get keys, because that will cause test
	// failures 5 years from now :-D
	testTime := getCurrentTimeForTest(sigchain, keyFamily)
	numSibkeys := len(ckf.GetAllActiveSibkeysAtTime(testTime))
	if numSibkeys != testCase.Sibkeys {
		t.Fatalf("Expected %d sibkeys, got %d", testCase.Sibkeys, numSibkeys)
	}
	numSubkeys := len(ckf.GetAllActiveSubkeysAtTime(testTime))
	if numSubkeys != testCase.Subkeys {
		t.Fatalf("Expected %d subkeys, got %d", testCase.Subkeys, numSubkeys)
	}

	// Success!
}

func createKeyFamily(bundles []string) (*KeyFamily, error) {
	allKeys := jsonw.NewArray(len(bundles))
	for i, bundle := range bundles {
		err := allKeys.SetIndex(i, jsonw.NewString(bundle))
		if err != nil {
			return nil, err
		}
	}
	publicKeys := jsonw.NewDictionary()
	publicKeys.SetKey("all_bundles", allKeys)
	return ParseKeyFamily(publicKeys)
}

func getCurrentTimeForTest(sigChain SigChain, keyFamily *KeyFamily) time.Time {
	// Pick a test time that's the latest ctime of all links and PGP keys.
	var t time.Time
	for _, link := range sigChain.chainLinks {
		linkCTime := time.Unix(link.unpacked.ctime, 0)
		if linkCTime.After(t) {
			t = linkCTime
		}
	}
	for _, ks := range keyFamily.PGPKeySets {
		keyCTime := ks.PermissivelyMergedKey.PrimaryKey.CreationTime
		if keyCTime.After(t) {
			t = keyCTime
		}
	}
	return t
}
