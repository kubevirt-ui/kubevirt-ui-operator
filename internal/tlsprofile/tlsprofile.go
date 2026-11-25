// @ai-rules:
// 1. [Pattern]: Pure helper. No client writes; the only I/O is a single Get of apiservers.config.openshift.io/cluster.
// 2. [Constraint]: nginx output (NginxSSLProtocols/NginxSSLCiphers) is built ONLY from cipher names validated against
//    the union of names in configv1.TLSProfiles. Never interpolate unvalidated user input (nginx directive injection).
// 3. [Constraint]: When MinVersion is TLS 1.3, OpenSSLCiphers and CipherSuites are nil (TLS 1.3 suites are not configurable).
// 4. [Gotcha]: Holder.TLSOpt must clone the base config at handshake time so GetCertificate/Certificates that
//    controller-runtime adds after TLSOpts are preserved. Do not set GetCertificate here.

// Package tlsprofile derives TLS settings for the operator and its operands from the cluster-wide
// apiservers.config.openshift.io/cluster tlsSecurityProfile.
package tlsprofile

import (
	"context"
	"crypto/tls"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	configv1 "github.com/openshift/api/config/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// APIServerName is the name of the singleton APIServer config object.
const APIServerName = "cluster"

// Settings is the resolved, validated TLS configuration.
type Settings struct {
	// MinTLSVersion is the normalized OpenShift minimum TLS version; never empty.
	MinTLSVersion configv1.TLSProtocolVersion
	// MinVersion is MinTLSVersion as a crypto/tls version constant.
	MinVersion uint16
	// OpenSSLCiphers are the validated TLS<=1.2 OpenSSL cipher names in profile order (used by nginx).
	// Nil when MinVersion is TLS 1.3.
	OpenSSLCiphers []string
	// CipherSuites is the subset of OpenSSLCiphers supported by Go, as IANA IDs (used by the proxy and
	// operator metrics server). Nil when MinVersion is TLS 1.3.
	CipherSuites []uint16
}

var tls13CipherNames = map[string]struct{}{
	"TLS_AES_128_GCM_SHA256":       {},
	"TLS_AES_256_GCM_SHA384":       {},
	"TLS_CHACHA20_POLY1305_SHA256": {},
}

// openSSLToIANA maps OpenSSL cipher names to the IANA names used by crypto/tls.
// Names absent from this table have no Go implementation and are only emitted for nginx.
var openSSLToIANA = map[string]string{
	"ECDHE-ECDSA-AES128-GCM-SHA256": "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
	"ECDHE-RSA-AES128-GCM-SHA256":   "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
	"ECDHE-ECDSA-AES256-GCM-SHA384": "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
	"ECDHE-RSA-AES256-GCM-SHA384":   "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
	"ECDHE-ECDSA-CHACHA20-POLY1305": "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
	"ECDHE-RSA-CHACHA20-POLY1305":   "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
	"ECDHE-ECDSA-AES128-SHA256":     "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
	"ECDHE-RSA-AES128-SHA256":       "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
	"ECDHE-ECDSA-AES128-SHA":        "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
	"ECDHE-RSA-AES128-SHA":          "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
	"ECDHE-ECDSA-AES256-SHA":        "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
	"ECDHE-RSA-AES256-SHA":          "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
	"AES128-GCM-SHA256":             "TLS_RSA_WITH_AES_128_GCM_SHA256",
	"AES256-GCM-SHA384":             "TLS_RSA_WITH_AES_256_GCM_SHA384",
	"AES128-SHA256":                 "TLS_RSA_WITH_AES_128_CBC_SHA256",
	"AES128-SHA":                    "TLS_RSA_WITH_AES_128_CBC_SHA",
	"AES256-SHA":                    "TLS_RSA_WITH_AES_256_CBC_SHA",
	"DES-CBC3-SHA":                  "TLS_RSA_WITH_3DES_EDE_CBC_SHA",
}

var (
	allowedOnce    sync.Once
	allowedCiphers map[string]struct{} // union of all names in configv1.TLSProfiles
	suiteIDByName  map[string]uint16
)

func initTables() {
	allowedOnce.Do(func() {
		allowedCiphers = map[string]struct{}{}
		for _, spec := range configv1.TLSProfiles {
			for _, c := range spec.Ciphers {
				allowedCiphers[c] = struct{}{}
			}
		}
		suiteIDByName = map[string]uint16{}
		for _, s := range tls.CipherSuites() {
			suiteIDByName[s.Name] = s.ID
		}
		for _, s := range tls.InsecureCipherSuites() {
			suiteIDByName[s.Name] = s.ID
		}
	})
}

func versionOf(v configv1.TLSProtocolVersion) (uint16, bool) {
	switch v {
	case configv1.VersionTLS10:
		return tls.VersionTLS10, true
	case configv1.VersionTLS11:
		return tls.VersionTLS11, true
	case configv1.VersionTLS12:
		return tls.VersionTLS12, true
	case configv1.VersionTLS13:
		return tls.VersionTLS13, true
	}
	return 0, false
}

// Resolve converts a cluster TLSSecurityProfile into validated Settings. A nil profile, an empty or unknown
// type, or a Custom type without a custom spec resolves to the Intermediate profile.
func Resolve(p *configv1.TLSSecurityProfile) Settings {
	intermediate := configv1.TLSProfiles[configv1.TLSProfileIntermediateType]
	spec := intermediate
	if p != nil {
		switch p.Type {
		case configv1.TLSProfileOldType, configv1.TLSProfileIntermediateType, configv1.TLSProfileModernType:
			spec = configv1.TLSProfiles[p.Type]
		case configv1.TLSProfileCustomType:
			if p.Custom != nil {
				spec = &p.Custom.TLSProfileSpec
			}
		}
	}

	minTLS := spec.MinTLSVersion
	minVersion, ok := versionOf(minTLS)
	if !ok {
		minTLS, minVersion = configv1.VersionTLS12, tls.VersionTLS12
	}
	s := Settings{MinTLSVersion: minTLS, MinVersion: minVersion}
	if minVersion >= tls.VersionTLS13 {
		return s
	}

	initTables()
	seen := map[string]struct{}{}
	for _, c := range spec.Ciphers {
		if _, ok := tls13CipherNames[c]; ok {
			continue
		}
		if _, ok := allowedCiphers[c]; !ok {
			continue
		}
		if _, dup := seen[c]; dup {
			continue
		}
		seen[c] = struct{}{}
		s.OpenSSLCiphers = append(s.OpenSSLCiphers, c)
		if iana, ok := openSSLToIANA[c]; ok {
			if id, ok := suiteIDByName[iana]; ok {
				s.CipherSuites = append(s.CipherSuites, id)
			}
		}
	}
	return s
}

// FromAPIServer reads the cluster APIServer object and resolves its TLS profile.
// A missing object resolves to the default (Intermediate) profile.
func FromAPIServer(ctx context.Context, c client.Reader) (Settings, error) {
	api := &configv1.APIServer{}
	if err := c.Get(ctx, client.ObjectKey{Name: APIServerName}, api); err != nil {
		if apierrors.IsNotFound(err) {
			return Resolve(nil), nil
		}
		return Settings{}, err
	}
	return Resolve(api.Spec.TLSSecurityProfile), nil
}

// ProxyArgs returns the flags for the apiserver proxy container.
func (s Settings) ProxyArgs() []string {
	args := []string{"--tls-min-version=" + strconv.Itoa(int(s.MinVersion))}
	if s.MinVersion < tls.VersionTLS13 && len(s.CipherSuites) > 0 {
		ids := make([]string, len(s.CipherSuites))
		for i, id := range s.CipherSuites {
			ids[i] = strconv.Itoa(int(id))
		}
		args = append(args, "--tls-cipher-suites="+strings.Join(ids, ","))
	}
	return args
}

// NginxSSLProtocols returns the value for the nginx ssl_protocols directive.
func (s Settings) NginxSSLProtocols() string {
	all := []struct {
		v    uint16
		name string
	}{
		{tls.VersionTLS10, "TLSv1"},
		{tls.VersionTLS11, "TLSv1.1"},
		{tls.VersionTLS12, "TLSv1.2"},
		{tls.VersionTLS13, "TLSv1.3"},
	}
	var out []string
	for _, p := range all {
		if p.v >= s.MinVersion {
			out = append(out, p.name)
		}
	}
	return strings.Join(out, " ")
}

// NginxSSLCiphers returns the value for the nginx ssl_ciphers directive; empty means omit the directive.
func (s Settings) NginxSSLCiphers() string {
	if s.MinVersion >= tls.VersionTLS13 {
		return ""
	}
	return strings.Join(s.OpenSSLCiphers, ":")
}

// Apply sets the minimum version and cipher suites on cfg.
func (s Settings) Apply(cfg *tls.Config) {
	cfg.MinVersion = s.MinVersion
	if s.MinVersion < tls.VersionTLS13 && len(s.CipherSuites) > 0 {
		cfg.CipherSuites = slices.Clone(s.CipherSuites)
	} else {
		cfg.CipherSuites = nil
	}
}

// Holder carries the current Settings so a long-lived listener can pick up profile changes without a restart.
type Holder struct {
	cur atomic.Pointer[Settings]
}

// NewHolder returns a Holder initialised with s.
func NewHolder(s Settings) *Holder {
	h := &Holder{}
	h.Store(s)
	return h
}

// Store replaces the current Settings.
func (h *Holder) Store(s Settings) { h.cur.Store(&s) }

// Load returns the current Settings (the default profile if none has been stored).
func (h *Holder) Load() Settings {
	if p := h.cur.Load(); p != nil {
		return *p
	}
	return Resolve(nil)
}

// TLSOpt returns a controller-runtime TLSOpts function that applies the current Settings on every handshake.
func (h *Holder) TLSOpt() func(*tls.Config) {
	return func(base *tls.Config) {
		base.GetConfigForClient = func(*tls.ClientHelloInfo) (*tls.Config, error) {
			c := base.Clone()
			c.GetConfigForClient = nil
			h.Load().Apply(c)
			return c, nil
		}
	}
}
