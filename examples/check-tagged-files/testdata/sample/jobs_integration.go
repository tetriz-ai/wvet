//go:build integration

package sample

// StartFixture starts a fixture server for the integration tests.
func StartFixture(serve func()) {
	go serve()
}
