// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import "database/sql"

// Minimal mocks of the embedded field types.

type fakeDB struct{}

func (f *fakeDB) Query(query string, args ...interface{}) (*sql.Rows, error) { return nil, nil }
func (f *fakeDB) QueryRow(query string, args ...interface{}) *sql.Row        { return nil }
func (f *fakeDB) Get(dest interface{}, query string, args ...interface{}) error {
	return nil
}
func (f *fakeDB) Select(dest interface{}, query string, args ...interface{}) error { return nil }
func (f *fakeDB) Exec(query string, args ...interface{}) (sql.Result, error)       { return nil, nil }

type fakeTx struct{}

func (f *fakeTx) Query(query string, args ...interface{}) (*sql.Rows, error) { return nil, nil }
func (f *fakeTx) QueryRow(query string, args ...interface{}) *sql.Row        { return nil }
func (f *fakeTx) Exec(query string, args ...interface{}) (sql.Result, error) { return nil, nil }

// Wrapper types — names must match what the analyzer checks.

type sqlxDBWrapper struct {
	DB *fakeDB
}

func (w *sqlxDBWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) { return nil, nil }
func (w *sqlxDBWrapper) QueryRow(query string, args ...interface{}) *sql.Row        { return nil }
func (w *sqlxDBWrapper) Get(dest interface{}, query string, args ...interface{}) error {
	return nil
}
func (w *sqlxDBWrapper) Select(dest interface{}, query string, args ...interface{}) error {
	return nil
}
func (w *sqlxDBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) { return nil, nil }

type sqlxTxWrapper struct {
	Tx *fakeTx
}

func (w *sqlxTxWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) { return nil, nil }
func (w *sqlxTxWrapper) QueryRow(query string, args ...interface{}) *sql.Row        { return nil }
func (w *sqlxTxWrapper) Exec(query string, args ...interface{}) (sql.Result, error) { return nil, nil }

// --- Tests ---

func validUsage(dbw *sqlxDBWrapper, txw *sqlxTxWrapper) {
	// Calling methods on the wrapper itself is always fine.
	_, _ = dbw.Query("SELECT 1")
	_ = dbw.QueryRow("SELECT 1")
	_ = dbw.Get(nil, "SELECT 1")
	_ = dbw.Select(nil, "SELECT 1")
	_, _ = dbw.Exec("DELETE FROM Foo WHERE Id = ?", "x")

	_, _ = txw.Query("SELECT 1")
	_ = txw.QueryRow("SELECT 1")
	_, _ = txw.Exec("DELETE FROM Foo WHERE Id = ?", "x")
}

func invalidDirectDBAccess(dbw *sqlxDBWrapper) {
	_, _ = dbw.DB.Query("SELECT 1")    // want `direct call to \.DB\.Query\(\)`
	_ = dbw.DB.QueryRow("SELECT 1")   // want `direct call to \.DB\.QueryRow\(\)`
	_ = dbw.DB.Get(nil, "SELECT 1")   // want `direct call to \.DB\.Get\(\)`
	_ = dbw.DB.Select(nil, "SELECT 1") // want `direct call to \.DB\.Select\(\)`
	_, _ = dbw.DB.Exec("DELETE FROM Foo WHERE Id = ?", "x") // want `direct call to \.DB\.Exec\(\)`
}

func invalidDirectTxAccess(txw *sqlxTxWrapper) {
	_, _ = txw.Tx.Query("SELECT 1")  // want `direct call to \.Tx\.Query\(\)`
	_ = txw.Tx.QueryRow("SELECT 1") // want `direct call to \.Tx\.QueryRow\(\)`
	_, _ = txw.Tx.Exec("DELETE FROM Foo WHERE Id = ?", "x") // want `direct call to \.Tx\.Exec\(\)`
}

// Unrelated type with a DB field — must not be flagged.
type otherType struct {
	DB *fakeDB
}

func validUnrelatedType(o *otherType) {
	_, _ = o.DB.Query("SELECT 1")
}
