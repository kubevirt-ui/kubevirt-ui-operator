package builders

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/labels"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/names"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/tlsprofile"
)

// nginxConfig renders nginx.conf. ssl_protocols and ssl_ciphers come from s, whose cipher names are validated
// against the OpenShift TLS profile allow-list; ssl_ciphers is omitted when s has no TLS<=1.2 ciphers (TLS 1.3 only).
func nginxConfig(port int32, s tlsprofile.Settings) string {
	sslCiphers := ""
	if c := s.NginxSSLCiphers(); c != "" {
		sslCiphers = fmt.Sprintf("    ssl_ciphers         %s;\n", c)
	}
	return fmt.Sprintf(`error_log /dev/stdout info;
events {}
http {
  access_log         /dev/stdout;
  include            /etc/nginx/mime.types;
  default_type       application/octet-stream;
  keepalive_timeout  65;
  add_header X-Content-Type-Options nosniff;
  server {
    listen              %d ssl;
    listen              [::]:%d ssl;
    ssl_certificate     /var/serving-cert/tls.crt;
    ssl_certificate_key /var/serving-cert/tls.key;
    ssl_protocols       %s;
%s    root                /usr/share/nginx/html;

    location = /plugin-manifest.json {
      add_header Cache-Control 'no-cache, no-store, must-revalidate, proxy-revalidate, max-age=0';
      add_header Pragma 'no-cache';
      add_header Expires '0';
    }
    location = /plugin-entry.js {
      add_header Cache-Control 'no-cache, no-store, must-revalidate, proxy-revalidate, max-age=0';
      add_header Pragma 'no-cache';
      add_header Expires '0';
    }
  }
}
`, port, port, s.NginxSSLProtocols(), sslCiphers)
}

// nginxConfigHash is the hex sha256 of the rendered nginx.conf, used to roll plugin pods when it changes.
func nginxConfigHash(m *Context) string {
	sum := sha256.Sum256([]byte(nginxConfig(names.PluginPort, m.TLS)))
	return hex.EncodeToString(sum[:])
}

func NginxConfigMap(m *Context) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		Name:      names.NginxConfigMap,
		Namespace: m.OperandNamespace,
		Labels:    labels.Plugin(m.CR),
		Data: map[string]string{
			"nginx.conf": nginxConfig(names.PluginPort, m.TLS),
		},
	}
}

func UserSettingsConfigMap(m *Context) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		Name:      names.UserSettingsConfigMap,
		Namespace: m.OperandNamespace,
		Labels:    labels.UIConfig(m.CR),
		Data:      map[string]string{},
	}
}

func FeaturesConfigMapShell(m *Context, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		Name:      names.FeaturesConfigMap,
		Namespace: m.OperandNamespace,
		Labels:    labels.UIConfig(m.CR),
		Data:      data,
	}
}
