package builders

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	configv1 "github.com/openshift/api/config/v1"

	consoleuiv1alpha1 "github.com/kubevirt-ui/kubevirt-ui-operator/api/v1alpha1"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/names"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/operconfig"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/tlsprofile"
)

func testContext(p *configv1.TLSSecurityProfile) *Context {
	cr := &consoleuiv1alpha1.KubeVirtConsoleUI{}
	cr.Name = names.ResourceName
	cfg := operconfig.Config{OperandNamespace: "kubevirt-ui", PluginImage: "plugin:test", ProxyImage: "proxy:test"}
	return NewContext(cr, cfg, tlsprofile.Resolve(p))
}

var (
	modernProfile       = &configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType}
	intermediateProfile = &configv1.TLSSecurityProfile{Type: configv1.TLSProfileIntermediateType}
)

func TestProxyDeploymentTLSArgs(t *testing.T) {
	cases := []struct {
		name string
		p    *configv1.TLSSecurityProfile
		want []string
	}{
		{"intermediate", intermediateProfile, []string{
			"--tls-min-version=771",
			"--tls-cipher-suites=49195,49199,49196,49200,52393,52392",
		}},
		{"default", nil, []string{
			"--tls-min-version=771",
			"--tls-cipher-suites=49195,49199,49196,49200,52393,52392",
		}},
		{"modern", modernProfile, []string{"--tls-min-version=772"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := ProxyDeployment(testContext(tc.p))
			got := d.Spec.Template.Spec.Containers[0].Args
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("proxy args = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNginxConfigTLS(t *testing.T) {
	inter := NginxConfigMap(testContext(intermediateProfile)).Data["nginx.conf"]
	if !strings.Contains(inter, "ssl_protocols       TLSv1.2 TLSv1.3;") {
		t.Errorf("intermediate: missing ssl_protocols TLSv1.2 TLSv1.3:\n%s", inter)
	}
	wantCiphers := "ssl_ciphers         ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:" +
		"ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:" +
		"ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384;"
	if !strings.Contains(inter, wantCiphers) {
		t.Errorf("intermediate: missing ssl_ciphers line:\n%s", inter)
	}

	modern := NginxConfigMap(testContext(modernProfile)).Data["nginx.conf"]
	if !strings.Contains(modern, "ssl_protocols       TLSv1.3;") {
		t.Errorf("modern: missing ssl_protocols TLSv1.3:\n%s", modern)
	}
	if strings.Contains(modern, "ssl_ciphers") {
		t.Errorf("modern: ssl_ciphers must be omitted:\n%s", modern)
	}
}

func TestPluginNginxConfigHashAnnotation(t *testing.T) {
	hashOf := func(p *configv1.TLSSecurityProfile) (annotation, expected string) {
		m := testContext(p)
		conf := NginxConfigMap(m).Data["nginx.conf"]
		sum := sha256.Sum256([]byte(conf))
		d := PluginDeployment(m)
		return d.Spec.Template.Annotations[names.NginxConfigHashAnnotation], hex.EncodeToString(sum[:])
	}

	interAnn, interWant := hashOf(intermediateProfile)
	if interAnn == "" || interAnn != interWant {
		t.Errorf("intermediate: annotation %q, want sha256 of nginx.conf %q", interAnn, interWant)
	}
	modernAnn, modernWant := hashOf(modernProfile)
	if modernAnn != modernWant {
		t.Errorf("modern: annotation %q, want %q", modernAnn, modernWant)
	}
	if interAnn == modernAnn {
		t.Error("annotation must change when the TLS profile changes")
	}
	if got := PluginDeployment(testContext(intermediateProfile)).Spec.Template.Annotations["openshift.io/required-scc"]; got != "restricted-v2" {
		t.Errorf("required-scc annotation lost: %q", got)
	}
}
