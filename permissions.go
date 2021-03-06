package main

import (
	"net/http"

	"github.com/DHowett/ghostbin/model"
	"github.com/gorilla/sessions"
)

type globalPermissionScope struct {
	pID model.PasteID
	// attempts to merge session store with user perms
	// in the new unified scope interface.

	// User's paste perm scope for this ID
	uScope model.PermissionScope

	v3Entries map[model.PasteID]model.Permission
}

func (g *globalPermissionScope) Has(p model.Permission) bool {
	if g.uScope != nil {
		return g.uScope.Has(p)
	}
	return g.v3Entries[g.pID]&p == p
}

func (g *globalPermissionScope) Grant(p model.Permission) error {
	if g.uScope != nil {
		return g.uScope.Grant(p)
	}
	g.v3Entries[g.pID] = g.v3Entries[g.pID] | p
	return nil
}

func (g *globalPermissionScope) Revoke(p model.Permission) error {
	if g.uScope != nil {
		return g.uScope.Revoke(p)
	}
	g.v3Entries[g.pID] = g.v3Entries[g.pID] & (^p)
	if g.v3Entries[g.pID] == 0 {
		delete(g.v3Entries, g.pID)
	}
	return nil
}

func GetPastePermissionScope(pID model.PasteID, r *http.Request) model.PermissionScope {
	var userScope model.PermissionScope
	user := GetUser(r)
	if user != nil {
		userScope = user.Permissions(model.PermissionClassPaste, pID)
	}

	cookieSession, _ := sessionStore.Get(r, "session")
	v3EntriesI := cookieSession.Values["v3permissions"]
	v3Entries, ok := v3EntriesI.(map[model.PasteID]model.Permission)
	if !ok || v3Entries == nil {
		v3Entries = make(map[model.PasteID]model.Permission)
		cookieSession.Values["v3permissions"] = v3Entries
	}

	return &globalPermissionScope{
		pID:       pID,
		uScope:    userScope,
		v3Entries: v3Entries,
	}
}

func SavePastePermissionScope(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		sessions.Save(r, w)
	}
}
