package main

import "C"
import (
	"encoding/base64"
	"encoding/json"

	"decred.org/dcrwallet/v4/wallet/udb"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
)

//export currentReceiveAddress
func currentReceiveAddress(cName *C.char) *C.char {
	w, ok := loadedWallet(cName)
	if !ok {
		return errCResponse("wallet with name %q is not loaded", goString(cName))
	}

	// Don't return an address if not synced!
	synced, _ := w.IsSynced(w.ctx)
	if !synced {
		return errCResponseWithCode(ErrCodeNotSynced, "currentReceiveAddress requested on an unsynced wallet")
	}

	addr, err := w.CurrentAddress(udb.DefaultAccountNum)
	if err != nil {
		return errCResponse("w.CurrentAddress error: %v", err)
	}

	return successCResponse(addr.String())
}

//export newExternalAddress
func newExternalAddress(cName *C.char) *C.char {
	w, ok := loadedWallet(cName)
	if !ok {
		return errCResponse("wallet with name %q is not loaded", goString(cName))
	}

	// Don't return an address if not synced!
	synced, _ := w.IsSynced(w.ctx)
	if !synced {
		return errCResponseWithCode(ErrCodeNotSynced, "newExternalAddress requested on an unsynced wallet")
	}

	_, err := w.NewExternalAddress(w.ctx, udb.DefaultAccountNum)
	if err != nil {
		return errCResponse("w.NewExternalAddress error: %v", err)
	}

	// NewExternalAddress will take the current address before increasing
	// the index. Get the current address after increasing the index.
	addr, err := w.CurrentAddress(udb.DefaultAccountNum)
	if err != nil {
		return errCResponse("w.CurrentAddress error: %v", err)
	}

	return successCResponse(addr.String())
}

//export signMessage
func signMessage(cName, cMessage, cAddress, cPassword *C.char) *C.char {
	w, ok := loadedWallet(cName)
	if !ok {
		return errCResponse("wallet with name %q is not loaded", goString(cName))
	}

	addr, err := stdaddr.DecodeAddress(goString(cAddress), w.MainWallet().ChainParams())
	if err != nil {
		return errCResponse("unable to decode address: %v", err)
	}

	if err := w.MainWallet().Unlock(w.ctx, []byte(goString(cPassword)), nil); err != nil {
		return errCResponse("cannot unlock wallet: %v", err)
	}

	sig, err := w.MainWallet().SignMessage(w.ctx, goString(cMessage), addr)
	if err != nil {
		return errCResponse("unable to sign message: %v", err)
	}

	sEnc := base64.StdEncoding.EncodeToString(sig)

	return successCResponse(sEnc)
}

//export addresses
func addresses(cName *C.char) *C.char {
	w, ok := loadedWallet(cName)
	if !ok {
		return errCResponse("wallet with name %q is not loaded", goString(cName))
	}

	addrs, err := w.AddressesByAccount(w.ctx, defaultAccount)
	if err != nil {
		return errCResponse("w.AddressesByAccount error: %v", err)
	}

	// w.AddressesByAccount does not include the current address.
	synced, _ := w.IsSynced(w.ctx)
	if synced {
		addr, err := w.CurrentAddress(udb.DefaultAccountNum)
		if err != nil {
			return errCResponse("w.CurrentAddress error: %v", err)
		}
		addrs = append(addrs, addr.String())
	}

	b, err := json.Marshal(addrs)
	if err != nil {
		return errCResponse("unable to marshal addresses: %v", err)
	}

	return successCResponse(string(b))
}

//export defaultPubkey
func defaultPubkey(cName *C.char) *C.char {
	w, ok := loadedWallet(cName)
	if !ok {
		return errCResponse("wallet with name %q is not loaded", goString(cName))
	}

	pubkey, err := w.AccountPubkey(w.ctx, defaultAccount)
	if err != nil {
		return errCResponse("unable to get default pubkey: %v", err)
	}

	return successCResponse(pubkey)
}
