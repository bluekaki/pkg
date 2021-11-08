package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bluekaki/pkg/errors"
)

const (
	// DefaultTTL default signature ttl
	DefaultTTL = time.Minute
	// IdentifierLen identifier's fix length
	IdentifierLen = 6
	delimiter     = "|"
)

const (
	// GRPC the grpc method
	GRPC = "GRPC"
)

var _ Method = (*method)(nil)

// Method define supproted http method
type Method interface {
	String() string
	Unknow() bool
	t()
}

type method struct {
	value string
}

func (m *method) String() string { return m.value }
func (m *method) Unknow() bool   { return m.value == "UNKNOW" }
func (m *method) t()             {}

// Identifier distinguish what system is displayed, fix length is IdentifierLen
type Identifier = string

// Secret symmetric encryption cipher code
type Secret = string

var (
	// MethodUnknow unknow
	MethodUnknow Method = &method{value: "UNKNOW"}
	// MethodGet http get
	MethodGet Method = &method{value: http.MethodGet}
	// MethodHead http head
	MethodHead Method = &method{value: http.MethodHead}
	// MethodPost http post
	MethodPost Method = &method{value: http.MethodPost}
	// MethodPut http put
	MethodPut Method = &method{value: http.MethodPut}
	// MethodPatch http patch
	MethodPatch Method = &method{value: http.MethodPatch}
	// MethodDelete http delete
	MethodDelete Method = &method{value: http.MethodDelete}
	// MethodConnect http connect
	MethodConnect Method = &method{value: http.MethodConnect}
	// MethodOptions http options
	MethodOptions Method = &method{value: http.MethodOptions}
	// MethodTrace http trace
	MethodTrace Method = &method{value: http.MethodTrace}
	// MethodGRPC grpc
	MethodGRPC Method = &method{value: GRPC}
)

// ToMethod convert to method
func ToMethod(method string) Method {
	switch strings.ToUpper(method) {
	case http.MethodGet:
		return MethodGet
	case http.MethodHead:
		return MethodHead
	case http.MethodPost:
		return MethodPost
	case http.MethodPut:
		return MethodPut
	case http.MethodPatch:
		return MethodPatch
	case http.MethodDelete:
		return MethodDelete
	case http.MethodConnect:
		return MethodConnect
	case http.MethodOptions:
		return MethodOptions
	case http.MethodTrace:
		return MethodTrace
	case GRPC:
		return MethodGRPC
	default:
		return MethodUnknow
	}
}

var _ Signature = (*signature)(nil)

// Signature defines methods of signature
type Signature interface {
	ResetSecrets(secrets map[Identifier]Secret) error
	Generate(identifier Identifier, method Method, uri string, body []byte) (authorization, date string, err error)
	Verify(authorization, date string, method Method, uri string, body []byte) (identifier Identifier, ok bool, err error)
}

// Option optional config
type Option func(*option)

type option struct {
	authorizationLen int
	hash             func() hash.Hash
	ttl              time.Duration
	secrets          map[Identifier]Secret
}

type signature struct {
	mux              sync.RWMutex
	authorizationLen int
	hash             func() hash.Hash
	ttlSeconds       float64
	secrets          map[Identifier]Secret
}

func genAuthorizationLen(hashSize int) int {
	// TODO md5/sha1/sha256 pass, other algorithms need to be verified
	return IdentifierLen + 1 + int(math.Ceil(math.Ceil(float64(hashSize)/3)*4)) // len(identifier) + 1(space) + base64(hashSize)
}

// WithMD5 use md5 hash algorithm
func WithMD5() Option {
	return func(opt *option) {
		opt.authorizationLen = genAuthorizationLen(md5.Size)
		opt.hash = md5.New
	}
}

// WithSHA1 use sha1 hash algorithm
func WithSHA1() Option {
	return func(opt *option) {
		opt.authorizationLen = genAuthorizationLen(sha1.Size)
		opt.hash = sha1.New
	}
}

// WithSHA256 use sha256 hash algorithm
func WithSHA256() Option {
	return func(opt *option) {
		opt.authorizationLen = genAuthorizationLen(sha256.Size)
		opt.hash = sha256.New
	}
}

// WithTTL setup signature's ttl
func WithTTL(ttl time.Duration) Option {
	return func(opt *option) {
		opt.ttl = ttl
	}
}

// WithSecrets setup mutli identifier-secret
func WithSecrets(secrets map[Identifier]Secret) Option {
	return func(opt *option) {
		opt.secrets = secrets
	}
}

// NewSignature create a new signature instance
func NewSignature(opts ...Option) (Signature, error) {
	opt := new(option)
	for _, f := range opts {
		f(opt)
	}

	if opt.hash == nil {
		return nil, errors.New("hash algorithm required")
	}

	secrets, err := verifySecrets(opt.secrets)
	if err != nil {
		return nil, err
	}

	ttl := opt.ttl
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	return &signature{
		authorizationLen: opt.authorizationLen,
		hash:             opt.hash,
		ttlSeconds:       float64(ttl / time.Second),
		secrets:          secrets,
	}, nil
}

func (s *signature) getSecret(identifier Identifier) (Secret, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	secret, ok := s.secrets[identifier]
	return secret, ok
}

func (s *signature) ResetSecrets(secrets map[Identifier]Secret) error {
	secrets, err := verifySecrets(secrets)
	if err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	s.secrets = secrets
	return nil
}

func (s *signature) Generate(identifier Identifier, method Method, uri string, body []byte) (authorization, date string, err error) {
	if identifier == "" {
		err = errors.New("identifier required")
		return
	}

	if method == nil {
		err = errors.New("method required")
		return
	}

	if uri == "" {
		err = errors.New("uri required")
		return
	}

	if decodedUri, err := url.QueryUnescape(uri); err == nil {
		uri = decodedUri
	}

	secret, ok := s.getSecret(identifier)
	if !ok {
		err = errors.Errorf("identifier %s not defined", identifier)
		return
	}

	date = time.Now().UTC().Format(http.TimeFormat)

	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(method.String())
	buffer.WriteString(delimiter)
	buffer.WriteString(uri)
	buffer.WriteString(delimiter)
	buffer.Write(body)
	buffer.WriteString(delimiter)
	buffer.WriteString(date)

	hash := hmac.New(s.hash, []byte(secret))
	hash.Write(buffer.Bytes())
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	authorization = fmt.Sprintf("%s %s", identifier, digest)
	return
}

func (s *signature) Verify(authorization, date string, method Method, uri string, body []byte) (identifier Identifier, ok bool, err error) {
	if len(authorization) != s.authorizationLen {
		err = errors.Errorf("authorization length must be %d", s.authorizationLen)
		return
	}

	if date == "" {
		err = errors.New("date required")
		return
	}

	if method == nil {
		err = errors.New("method required")
		return
	}

	if uri == "" {
		err = errors.New("uri required")
		return
	}

	if decodedUri, err := url.QueryUnescape(uri); err == nil {
		uri = decodedUri
	}

	ts, err := time.ParseInLocation(http.TimeFormat, date, time.UTC)
	if err != nil {
		err = errors.New("date must follow 'time.RFC1123 GMT' in format of 'DAY, DD MON YYYY hh:mm:ss GMT'")
		return
	}

	if math.Abs(time.Now().UTC().Sub(ts).Seconds()) > s.ttlSeconds {
		err = errors.Errorf("date exceeds limit %.f seconds", s.ttlSeconds)
		return
	}

	identifier = authorization[:IdentifierLen]
	secret, ok := s.getSecret(identifier)
	if !ok {
		err = errors.Errorf("identifier %s not supported", identifier)
		return
	}

	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(method.String())
	buffer.WriteString(delimiter)
	buffer.WriteString(uri)
	buffer.WriteString(delimiter)
	buffer.Write(body)
	buffer.WriteString(delimiter)
	buffer.WriteString(date)

	hash := hmac.New(s.hash, []byte(secret))
	hash.Write(buffer.Bytes())
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	ok = authorization[IdentifierLen+1:] == digest
	return
}

func verifySecrets(secrets map[Identifier]Secret) (map[Identifier]Secret, error) {
	if len(secrets) == 0 {
		return nil, errors.New("secrets required")
	}

	clone := make(map[Identifier]Secret, len(secrets))
	for identifier, secret := range secrets {
		identifier = strings.TrimSpace(identifier)
		if len(identifier) != len([]rune(identifier)) {
			return nil, errors.New("identifier must be ascii")
		}

		if len(identifier) != IdentifierLen {
			return nil, errors.Errorf("identifier length must be %d", IdentifierLen)
		}

		if secret = strings.TrimSpace(secret); secret == "" {
			return nil, errors.New("secret can not be empty")
		}

		clone[identifier] = secret
	}

	return clone, nil
}
