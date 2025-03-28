package hcloud

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

func TestCertificateCreateOptsValidate_Uploaded(t *testing.T) {
	tests := []struct {
		name   string
		opts   CertificateCreateOpts
		errMsg string
	}{
		{
			name: "missing name",
			opts: CertificateCreateOpts{
				Certificate: "cert",
				PrivateKey:  "key",
				Labels:      map[string]string{},
			},
			errMsg: "missing field [Name] in [hcloud.CertificateCreateOpts]",
		},
		{
			name: "no certificate",
			opts: CertificateCreateOpts{
				Name:       "name",
				PrivateKey: "key",
				Labels:     map[string]string{},
			},
			errMsg: "missing field [Certificate] in [hcloud.CertificateCreateOpts]",
		},
		{
			name: "no private key",
			opts: CertificateCreateOpts{
				Name:        "name",
				Certificate: "cert",
				Labels:      map[string]string{},
			},
			errMsg: "missing field [PrivateKey] in [hcloud.CertificateCreateOpts]",
		},
		{
			name: "valid without type",
			opts: CertificateCreateOpts{
				Name:        "name",
				Certificate: "cert",
				PrivateKey:  "key",
				Labels:      map[string]string{},
			},
		},
		{
			name: "valid with type",
			opts: CertificateCreateOpts{
				Name:        "name",
				Type:        "uploaded",
				Certificate: "cert",
				PrivateKey:  "key",
				Labels:      map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.errMsg != "" {
				assert.EqualError(t, err, tt.errMsg)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestCertificateCreateOptsValidate_Managed(t *testing.T) {
	tests := []struct {
		name   string
		opts   CertificateCreateOpts
		errMsg string
	}{
		{
			name: "missing name",
			opts: CertificateCreateOpts{
				Type:        "managed",
				DomainNames: []string{"*.example.com", "example.com"},
			},
			errMsg: "missing field [Name] in [hcloud.CertificateCreateOpts]",
		},
		{
			name: "missing domains",
			opts: CertificateCreateOpts{
				Name: "I have no domains",
				Type: "managed",
			},
			errMsg: "missing field [DomainNames] in [hcloud.CertificateCreateOpts]",
		},
		{
			name: "valid certificate",
			opts: CertificateCreateOpts{
				Name:        "valid",
				Type:        "managed",
				DomainNames: []string{"*.example.com", "example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.errMsg != "" {
				assert.EqualError(t, err, tt.errMsg)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestCertificateCreateOptsValidate_InvalidType(t *testing.T) {
	opts := CertificateCreateOpts{Name: "invalid type", Type: "invalid"}
	err := opts.Validate()
	assert.EqualError(t, err, "invalid value 'invalid' for field [Type] in [hcloud.CertificateCreateOpts]")
}

func TestCertificateClientGetByID(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/certificates/1", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(schema.CertificateGetResponse{
			Certificate: schema.Certificate{
				ID: 1,
			},
		})
	})

	ctx := context.Background()

	certficate, _, err := env.Client.Certificate.GetByID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if certficate == nil {
		t.Fatal("no certficate")
	}
	if certficate.ID != 1 {
		t.Errorf("unexpected certficate ID: %v", certficate.ID)
	}

	t.Run("called via Get", func(t *testing.T) {
		certficate, _, err := env.Client.Certificate.Get(ctx, "1")
		if err != nil {
			t.Fatal(err)
		}
		if certficate == nil {
			t.Fatal("no certficate")
		}
		if certficate.ID != 1 {
			t.Errorf("unexpected certficate ID: %v", certficate.ID)
		}
	})
}

func TestCertificateClientGetByIDNotFound(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/certificates/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(schema.ErrorResponse{
			Error: schema.Error{
				Code: string(ErrorCodeNotFound),
			},
		})
	})

	ctx := context.Background()

	certficate, _, err := env.Client.Certificate.GetByID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if certficate != nil {
		t.Fatal("expected no certficate")
	}
}

func TestCertificateClientGetByName(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/certificates", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "name=mycert" {
			t.Fatal("missing name query")
		}
		json.NewEncoder(w).Encode(schema.CertificateListResponse{
			Certificates: []schema.Certificate{
				{
					ID:   1,
					Name: "mycert",
				},
			},
		})
	})

	ctx := context.Background()

	certficate, _, err := env.Client.Certificate.GetByName(ctx, "mycert")
	if err != nil {
		t.Fatal(err)
	}
	if certficate == nil {
		t.Fatal("no certficate")
	}
	if certficate.ID != 1 {
		t.Errorf("unexpected certficate ID: %v", certficate.ID)
	}

	t.Run("via Get", func(t *testing.T) {
		certficate, _, err := env.Client.Certificate.Get(ctx, "mycert")
		if err != nil {
			t.Fatal(err)
		}
		if certficate == nil {
			t.Fatal("no certficate")
		}
		if certficate.ID != 1 {
			t.Errorf("unexpected certficate ID: %v", certficate.ID)
		}
	})
}

func TestCertificateClientGetByNumericName(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/certificates/123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(schema.ErrorResponse{
			Error: schema.Error{
				Code: string(ErrorCodeNotFound),
			},
		})
	})

	env.Mux.HandleFunc("/certificates", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "name=123" {
			t.Fatal("missing name query")
		}
		json.NewEncoder(w).Encode(schema.CertificateListResponse{
			Certificates: []schema.Certificate{
				{
					ID:   1,
					Name: "123",
				},
			},
		})
	})

	ctx := context.Background()

	certficate, _, err := env.Client.Certificate.Get(ctx, "123")
	if err != nil {
		t.Fatal(err)
	}
	if certficate == nil {
		t.Fatal("no certficate")
	}
	if certficate.ID != 1 {
		t.Errorf("unexpected certficate ID: %v", certficate.ID)
	}
}

func TestCertificateClientGetByNameNotFound(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/certificates", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "name=mycert" {
			t.Fatal("missing name query")
		}
		json.NewEncoder(w).Encode(schema.CertificateListResponse{
			Certificates: []schema.Certificate{},
		})
	})

	ctx := context.Background()

	certficate, _, err := env.Client.Certificate.GetByName(ctx, "mycert")
	if err != nil {
		t.Fatal(err)
	}
	if certficate != nil {
		t.Fatal("unexpected certficate")
	}
}

func TestCertificateClientGetByNameEmpty(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	ctx := context.Background()

	certficate, _, err := env.Client.Certificate.GetByName(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if certficate != nil {
		t.Fatal("unexpected certficate")
	}
}

func TestCertificateCreate_Uploaded(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/certificates", func(w http.ResponseWriter, r *http.Request) {
		var reqBody schema.CertificateCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatal(err)
		}
		expectedReqBody := schema.CertificateCreateRequest{
			Name:        "my-cert",
			Type:        "uploaded",
			Certificate: "-----BEGIN CERTIFICATE-----\n...",
			PrivateKey:  "-----BEGIN PRIVATE KEY-----\n...",
			Labels: func() *map[string]string {
				labels := map[string]string{"key": "value"}
				return &labels
			}(),
		}
		if !cmp.Equal(expectedReqBody, reqBody) {
			t.Log(cmp.Diff(expectedReqBody, reqBody))
			t.Error("unexpected request body")
		}
		json.NewEncoder(w).Encode(schema.CertificateCreateResponse{
			Certificate: schema.Certificate{ID: 1},
		})
	})

	ctx := context.Background()

	opts := CertificateCreateOpts{
		Name:        "my-cert",
		Certificate: "-----BEGIN CERTIFICATE-----\n...",
		PrivateKey:  "-----BEGIN PRIVATE KEY-----\n...",
		Labels:      map[string]string{"key": "value"},
	}
	_, _, err := env.Client.Certificate.Create(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCertificateCreate_Managed(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/certificates", func(w http.ResponseWriter, r *http.Request) {
		var reqBody schema.CertificateCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatal(err)
		}
		expectedReqBody := schema.CertificateCreateRequest{
			Name:        "my-cert",
			Type:        "managed",
			DomainNames: []string{"*.example.com", "example.com"},
			Labels: func() *map[string]string {
				labels := map[string]string{"key": "value"}
				return &labels
			}(),
		}
		if !cmp.Equal(expectedReqBody, reqBody) {
			t.Log(cmp.Diff(expectedReqBody, reqBody))
			t.Error("unexpected request body")
		}
		json.NewEncoder(w).Encode(schema.CertificateCreateResponse{
			Certificate: schema.Certificate{ID: 1},
			Action:      &schema.Action{ID: 14},
		})
	})

	ctx := context.Background()
	opts := CertificateCreateOpts{
		Name:        "my-cert",
		Type:        "managed",
		DomainNames: []string{"*.example.com", "example.com"},
		Labels:      map[string]string{"key": "value"},
	}

	result, resp, err := env.Client.Certificate.CreateCertificate(ctx, opts)
	assert.NoError(t, err)
	assert.NotNil(t, resp, "no response returned")
	assert.NotNil(t, result.Certificate, "no certificate returned")
	assert.NotNil(t, result.Action, "no action returned")
}

func TestCertificateCreateValidation(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	ctx := context.Background()
	opts := CertificateCreateOpts{}
	_, _, err := env.Client.Certificate.Create(ctx, opts)

	require.EqualError(t, err, "missing field [Name] in [hcloud.CertificateCreateOpts]")
}

func TestCertificateDelete(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/certificates/1", func(w http.ResponseWriter, r *http.Request) {})

	var (
		ctx        = context.Background()
		certficate = &Certificate{ID: 1}
	)

	_, err := env.Client.Certificate.Delete(ctx, certficate)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCertificateClientUpdate(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/certificates/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Error("expected PUT")
		}
		var reqBody schema.CertificateUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatal(err)
		}
		expectedReqBody := schema.CertificateUpdateRequest{
			Name: Ptr("test"),
			Labels: func() *map[string]string {
				labels := map[string]string{"key": "value"}
				return &labels
			}(),
		}
		if !cmp.Equal(expectedReqBody, reqBody) {
			t.Log(cmp.Diff(expectedReqBody, reqBody))
			t.Error("unexpected request body")
		}
		json.NewEncoder(w).Encode(schema.CertificateUpdateResponse{
			Certificate: schema.Certificate{
				ID: 1,
			},
		})
	})

	var (
		ctx        = context.Background()
		certficate = &Certificate{ID: 1}
	)

	opts := CertificateUpdateOpts{
		Name:   "test",
		Labels: map[string]string{"key": "value"},
	}
	updatedCertificate, _, err := env.Client.Certificate.Update(ctx, certficate, opts)
	if err != nil {
		t.Fatal(err)
	}
	if updatedCertificate.ID != 1 {
		t.Errorf("unexpected certficate ID: %v", updatedCertificate.ID)
	}
}

func TestCertificateClient_RetryIssuance(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/certificates/1/actions/retry", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		resp := schema.CertificateIssuanceRetryResponse{
			Action: schema.Action{ID: 1, Resources: []schema.ActionResourceReference{}},
		}
		err := json.NewEncoder(w).Encode(resp)
		assert.NoError(t, err)
	})

	action, _, err := env.Client.Certificate.RetryIssuance(context.Background(), &Certificate{ID: 1})
	assert.NoError(t, err)
	assert.Equal(t, &Action{ID: 1, Resources: []*ActionResource{}}, action)
}
