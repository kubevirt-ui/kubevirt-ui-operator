package tlsprofile

import (
	"context"
	"crypto/tls"
	"reflect"
	"strings"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const intermediateNginxCiphers = "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:" +
	"ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:" +
	"ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:" +
	"DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384"

func assertIntermediate(t *testing.T, s Settings) {
	t.Helper()
	if s.MinTLSVersion != configv1.VersionTLS12 || s.MinVersion != tls.VersionTLS12 {
		t.Fatalf("min version = %q/%d, want VersionTLS12/771", s.MinTLSVersion, s.MinVersion)
	}
	wantArgs := []string{"--tls-min-version=771", "--tls-cipher-suites=49195,49199,49196,49200,52393,52392"}
	if got := s.ProxyArgs(); !reflect.DeepEqual(got, wantArgs) {
		t.Errorf("ProxyArgs = %v, want %v", got, wantArgs)
	}
	if got := s.NginxSSLProtocols(); got != "TLSv1.2 TLSv1.3" {
		t.Errorf("NginxSSLProtocols = %q", got)
	}
	if got := s.NginxSSLCiphers(); got != intermediateNginxCiphers {
		t.Errorf("NginxSSLCiphers = %q", got)
	}
}

func TestResolveDefaultsToIntermediate(t *testing.T) {
	cases := map[string]*configv1.TLSSecurityProfile{
		"nil":          nil,
		"empty type":   {},
		"unknown type": {Type: "Bogus"},
		"custom nil":   {Type: configv1.TLSProfileCustomType},
		"intermediate": {Type: configv1.TLSProfileIntermediateType},
	}
	for name, p := range cases {
		t.Run(name, func(t *testing.T) { assertIntermediate(t, Resolve(p)) })
	}
}

func TestResolveModern(t *testing.T) {
	s := Resolve(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType})
	if s.MinVersion != 772 {
		t.Fatalf("MinVersion = %d", s.MinVersion)
	}
	if s.OpenSSLCiphers != nil || s.CipherSuites != nil {
		t.Errorf("TLS 1.3 must have nil ciphers: %v %v", s.OpenSSLCiphers, s.CipherSuites)
	}
	if got := s.ProxyArgs(); !reflect.DeepEqual(got, []string{"--tls-min-version=772"}) {
		t.Errorf("ProxyArgs = %v", got)
	}
	if got := s.NginxSSLProtocols(); got != "TLSv1.3" {
		t.Errorf("NginxSSLProtocols = %q", got)
	}
	if got := s.NginxSSLCiphers(); got != "" {
		t.Errorf("NginxSSLCiphers = %q, want empty", got)
	}
}

func TestResolveOld(t *testing.T) {
	s := Resolve(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileOldType})
	if s.MinVersion != 769 {
		t.Fatalf("MinVersion = %d", s.MinVersion)
	}
	if got := s.NginxSSLProtocols(); got != "TLSv1 TLSv1.1 TLSv1.2 TLSv1.3" {
		t.Errorf("NginxSSLProtocols = %q", got)
	}
	if len(s.CipherSuites) == 0 || len(s.CipherSuites) > len(s.OpenSSLCiphers) {
		t.Errorf("CipherSuites %d must be a non-empty subset of OpenSSLCiphers %d", len(s.CipherSuites), len(s.OpenSSLCiphers))
	}
	for _, c := range s.OpenSSLCiphers {
		if strings.HasPrefix(c, "TLS_") {
			t.Errorf("TLS 1.3 cipher %q leaked into OpenSSLCiphers", c)
		}
	}
	tls11 := Resolve(&configv1.TLSSecurityProfile{
		Type:   configv1.TLSProfileCustomType,
		Custom: &configv1.CustomTLSProfile{MinTLSVersion: configv1.VersionTLS11},
	})
	if got := tls11.NginxSSLProtocols(); got != "TLSv1.1 TLSv1.2 TLSv1.3" {
		t.Errorf("TLS1.1 NginxSSLProtocols = %q", got)
	}
}

func TestResolveCustom(t *testing.T) {
	s := Resolve(&configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileCustomType,
		Custom: &configv1.CustomTLSProfile{
			MinTLSVersion: configv1.VersionTLS12,
			Ciphers: []string{
				"ECDHE-RSA-AES128-GCM-SHA256", "bogus;evil", "TLS_AES_128_GCM_SHA256", "DHE-RSA-AES128-GCM-SHA256",
			}},
	})
	if want := []string{"ECDHE-RSA-AES128-GCM-SHA256", "DHE-RSA-AES128-GCM-SHA256"}; !reflect.DeepEqual(s.OpenSSLCiphers, want) {
		t.Errorf("OpenSSLCiphers = %v, want %v", s.OpenSSLCiphers, want)
	}
	if want := []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256}; !reflect.DeepEqual(s.CipherSuites, want) {
		t.Errorf("CipherSuites = %v, want %v (DHE is not supported by Go)", s.CipherSuites, want)
	}
	if got := s.NginxSSLCiphers(); got != "ECDHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES128-GCM-SHA256" {
		t.Errorf("NginxSSLCiphers = %q", got)
	}
}

func TestResolveCustomTLS13(t *testing.T) {
	s := Resolve(&configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileCustomType,
		Custom: &configv1.CustomTLSProfile{
			MinTLSVersion: configv1.VersionTLS13,
			Ciphers:       []string{"ECDHE-RSA-AES128-GCM-SHA256"}},
	})
	if s.MinVersion != tls.VersionTLS13 || s.OpenSSLCiphers != nil || s.CipherSuites != nil {
		t.Errorf("unexpected settings: %+v", s)
	}
}

func TestResolveCustomEmptyMinVersion(t *testing.T) {
	s := Resolve(&configv1.TLSSecurityProfile{
		Type:   configv1.TLSProfileCustomType,
		Custom: &configv1.CustomTLSProfile{MinTLSVersion: "", Ciphers: nil},
	})
	if s.MinTLSVersion != configv1.VersionTLS12 || s.MinVersion != tls.VersionTLS12 {
		t.Errorf("min version = %q/%d, want TLS 1.2", s.MinTLSVersion, s.MinVersion)
	}
	if got := s.NginxSSLCiphers(); got != "" {
		t.Errorf("NginxSSLCiphers = %q, want empty", got)
	}
	if got := s.ProxyArgs(); !reflect.DeepEqual(got, []string{"--tls-min-version=771"}) {
		t.Errorf("ProxyArgs = %v", got)
	}
}

func TestFromAPIServer(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := configv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	api := &configv1.APIServer{
		Name: APIServerName,
		Spec: configv1.APIServerSpec{
			TLSSecurityProfile: &configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType},
		},
	}
	s, err := FromAPIServer(ctx, fake.NewClientBuilder().WithScheme(scheme).WithObjects(api).Build())
	if err != nil {
		t.Fatal(err)
	}
	if s.MinVersion != tls.VersionTLS13 {
		t.Errorf("present: MinVersion = %d, want TLS 1.3", s.MinVersion)
	}

	s, err = FromAPIServer(ctx, fake.NewClientBuilder().WithScheme(scheme).Build())
	if err != nil {
		t.Fatalf("NotFound must not be an error: %v", err)
	}
	assertIntermediate(t, s)
}

func TestHolderTLSOpt(t *testing.T) {
	h := NewHolder(Resolve(nil))
	base := &tls.Config{NextProtos: []string{"h2", "http/1.1"}}
	h.TLSOpt()(base)
	// Mimic controller-runtime adding a certificate callback after TLSOpts have run.
	gotCert := false
	base.GetCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		gotCert = true
		return &tls.Certificate{}, nil
	}

	handshake := func() *tls.Config {
		t.Helper()
		c, err := base.GetConfigForClient(&tls.ClientHelloInfo{})
		if err != nil {
			t.Fatal(err)
		}
		return c
	}

	c := handshake()
	if c.MinVersion != tls.VersionTLS12 || len(c.CipherSuites) == 0 {
		t.Errorf("intermediate: MinVersion=%d suites=%d", c.MinVersion, len(c.CipherSuites))
	}
	if c.GetConfigForClient != nil {
		t.Error("GetConfigForClient must be cleared on the per-handshake config")
	}
	if !reflect.DeepEqual(c.NextProtos, []string{"h2", "http/1.1"}) {
		t.Errorf("NextProtos = %v", c.NextProtos)
	}
	if c.GetCertificate == nil {
		t.Fatal("GetCertificate set after TLSOpt was not preserved")
	}
	if _, err := c.GetCertificate(nil); err != nil || !gotCert {
		t.Errorf("GetCertificate not invoked correctly: err=%v called=%v", err, gotCert)
	}

	h.Store(Resolve(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType}))
	c = handshake()
	if c.MinVersion != tls.VersionTLS13 || c.CipherSuites != nil {
		t.Errorf("modern: MinVersion=%d suites=%v", c.MinVersion, c.CipherSuites)
	}
	if c.GetCertificate == nil || !reflect.DeepEqual(c.NextProtos, []string{"h2", "http/1.1"}) {
		t.Error("base settings lost after live update")
	}
	if base.MinVersion != 0 || base.CipherSuites != nil {
		t.Error("base config must not be mutated")
	}
}
