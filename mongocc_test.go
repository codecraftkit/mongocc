package mongocc

import (
	"strings"
	"testing"
)

func TestRedactMongoURI(t *testing.T) {
	tests := []struct {
		name     string
		mongoUri string
		want     string
	}{
		{
			name:     "credentials in userinfo are dropped",
			mongoUri: "mongodb://appuser:s3cr3t@cluster.example.com:27017/billing?retryWrites=true",
			want:     "mongodb://cluster.example.com:27017",
		},
		{
			name:     "srv scheme is preserved",
			mongoUri: "mongodb+srv://appuser:s3cr3t@cluster.abc.mongodb.net/billing?w=majority",
			want:     "mongodb+srv://cluster.abc.mongodb.net",
		},
		{
			name:     "every host of a replica set is kept",
			mongoUri: "mongodb://a.example:27017,b.example:27017/billing?replicaSet=rs0",
			want:     "mongodb://a.example:27017,b.example:27017",
		},
		{
			name:     "a uri without credentials is unchanged in substance",
			mongoUri: "mongodb://localhost:27017",
			want:     "mongodb://localhost:27017",
		},
		{
			name:     "an unescaped @ in the password does not confuse the host",
			mongoUri: "mongodb://appuser:p@ssw0rd@cluster.example.com:27017/billing",
			want:     "mongodb://cluster.example.com:27017",
		},
		{
			name:     "the tls key password lives in the option string, not the userinfo",
			mongoUri: "mongodb://cluster.example.com:27017/billing?tlsCertificateKeyFilePassword=s3cr3t",
			want:     "mongodb://cluster.example.com:27017",
		},
		{
			name:     "the aws session token lives in the option string too",
			mongoUri: "mongodb://cluster.example.com:27017/billing?authMechanismProperties=AWS_SESSION_TOKEN:s3cr3t",
			want:     "mongodb://cluster.example.com:27017",
		},
		{
			name:     "a string that does not parse is never echoed",
			mongoUri: "://///",
			want:     redactedURI,
		},
		{
			name:     "a string without a scheme is never echoed",
			mongoUri: "appuser:s3cr3t@cluster.example.com:27017",
			want:     redactedURI,
		},
		{
			name:     "a scheme without a host is never echoed",
			mongoUri: "mongodb:///billing",
			want:     redactedURI,
		},
		{
			name:     "the empty string is never echoed",
			mongoUri: "",
			want:     redactedURI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := redactMongoURI(tt.mongoUri); got != tt.want {
				t.Errorf("redactMongoURI(%q) = %q, want %q", tt.mongoUri, got, tt.want)
			}
		})
	}
}

// TestConnectionLogLineNeverLeaksSecrets is the reason this file exists.
//
// Until v1.3.0 Connect printed the whole connection string, so every service
// using mongocc wrote its MongoDB password to stdout on every boot, where
// container logs and log aggregators kept it. The table below carries a secret
// in each of the places a MongoDB connection string can hold one.
func TestConnectionLogLineNeverLeaksSecrets(t *testing.T) {
	const secret = "s3cr3t-do-not-log-me"

	uris := map[string]string{
		"password in userinfo":            "mongodb://appuser:" + secret + "@cluster.example.com:27017/billing",
		"password in userinfo over srv":   "mongodb+srv://appuser:" + secret + "@cluster.abc.mongodb.net/billing",
		"percent encoded password":        "mongodb://appuser:" + secret + "%40x@cluster.example.com:27017/billing",
		"tls certificate key password":    "mongodb://cluster.example.com:27017/billing?tlsCertificateKeyFilePassword=" + secret,
		"aws session token in properties": "mongodb://cluster.example.com:27017/b?authMechanismProperties=AWS_SESSION_TOKEN:" + secret,
		"unparseable but credentialed":    "://appuser:" + secret + "@cluster.example.com",
	}

	for name, mongoUri := range uris {
		t.Run(name, func(t *testing.T) {
			line := connectionLogLine(mongoUri, "billing")
			if strings.Contains(line, secret) {
				t.Errorf("connectionLogLine leaked the secret: %q", line)
			}
		})
	}
}

// TestConnectionLogLineKeepsDiagnosticValue guards the other direction. The
// message is worth keeping — it answers which server a service attached to —
// so redacting the whole line is not an acceptable way to pass the test above.
func TestConnectionLogLineKeepsDiagnosticValue(t *testing.T) {
	line := connectionLogLine("mongodb://appuser:s3cr3t@cluster.example.com:27017/billing", "billing_reports")

	for _, want := range []string{"cluster.example.com:27017", "billing_reports"} {
		if !strings.Contains(line, want) {
			t.Errorf("connectionLogLine dropped %q, leaving %q", want, line)
		}
	}
}
