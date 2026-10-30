// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package bait generates unique fake sensitive values per session and
// carrier (PLAN section 7). No generated value is ever committed; this
// package contains only formats and dictionaries.
package bait

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Class is a bait family (PLAN 7.2).
type Class string

// Bait classes.
const (
	ClassSecret   Class = "secret"
	ClassPassword Class = "password"
	ClassEmail    Class = "email"
	ClassPII      Class = "pii"
	ClassFact     Class = "fact"
)

// Subclass is a concrete bait format (PLAN 7.2).
type Subclass string

// Bait subclasses for v0.1.
const (
	OpenAILike      Subclass = "openai_like"
	AWSLike         Subclass = "aws_like"
	GitHubLike      Subclass = "github_like"
	JWT             Subclass = "jwt"
	DBDSN           Subclass = "db_dsn"
	WifiPassword    Subclass = "wifi_password"
	AccountPassword Subclass = "account_password"
	ResetToken      Subclass = "reset_token"
	EmailAddress    Subclass = "address"
	ThreadContent   Subclass = "thread_content"
	CCCD            Subclass = "cccd"
	Phone           Subclass = "phone"
	BankAccount     Subclass = "bank_account"
	Card            Subclass = "card"
	TaxID           Subclass = "tax_id"
	Address         Subclass = "address_pii"
	FictionalFact   Subclass = "fictional_fact"
)

var classOf = map[Subclass]Class{
	OpenAILike: ClassSecret, AWSLike: ClassSecret, GitHubLike: ClassSecret,
	JWT: ClassSecret, DBDSN: ClassSecret,
	WifiPassword: ClassPassword, AccountPassword: ClassPassword, ResetToken: ClassPassword,
	EmailAddress: ClassEmail, ThreadContent: ClassEmail,
	CCCD: ClassPII, Phone: ClassPII, BankAccount: ClassPII, Card: ClassPII,
	TaxID: ClassPII, Address: ClassPII, FictionalFact: ClassFact,
}

// Bait is one generated fake value. Value must never be logged in full.
type Bait struct {
	ID           string
	SessionID    string
	Class        Class
	Subclass     Subclass
	Value        string
	Identifier   string
	Fingerprints []string
	CarrierID    string
	ExpiresAt    time.Time
}

// Catalog returns every v0.1 subclass and its class (for
// `spillcheck bait preview`; formats only, never values).
func Catalog() map[Subclass]Class {
	out := make(map[Subclass]Class, len(classOf))
	for k, v := range classOf {
		out[k] = v
	}
	return out
}

// Generator creates baits deterministically from a seed (seed 42 is
// used by demo mode) and never repeats a value.
type Generator struct {
	rnd  *rand.Rand
	seed int64
	seen map[string]bool
	seq  int
}

// NewGenerator returns a generator for seed.
func NewGenerator(seed int64) *Generator {
	return &Generator{rnd: rand.New(rand.NewSource(seed)), seed: seed, seen: map[string]bool{}} //nolint:gosec // fake data, determinism required
}

// Generate creates one bait for sessionID/carrierID.
func (g *Generator) Generate(sessionID, carrierID string, sub Subclass) (Bait, error) {
	class, ok := classOf[sub]
	if !ok {
		return Bait{}, fmt.Errorf("bait: unknown subclass %q", sub)
	}
	for attempt := 0; attempt < 100; attempt++ {
		value, identifier := g.build(sessionID, carrierID, sub)
		if g.seen[value] {
			continue
		}
		g.seen[value] = true
		g.seq++
		return Bait{
			ID: fmt.Sprintf("bait-%d-%d", g.seed, g.seq), SessionID: sessionID,
			Class: class, Subclass: sub, Value: value, Identifier: identifier,
			Fingerprints: Fingerprints(value), CarrierID: carrierID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}, nil
	}
	return Bait{}, fmt.Errorf("bait: could not generate a unique %s bait", sub)
}

const base62 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
const upper62 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (g *Generator) chars(n int, alphabet string) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(alphabet[g.rnd.Intn(len(alphabet))])
	}
	return b.String()
}

func (g *Generator) digits(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(byte('0' + g.rnd.Intn(10)))
	}
	return b.String()
}

func (g *Generator) build(sessionID, carrierID string, sub Subclass) (string, string) {
	id := g.chars(6, base62)
	switch sub {
	case OpenAILike:
		return "sk-spill-" + g.chars(26, base62) + id, id
	case AWSLike:
		uid := strings.ToUpper(id)
		key := "ASPL" + g.chars(10, upper62) + uid
		return key + " " + g.chars(40, base62), uid
	case GitHubLike:
		return "ghp_spill_" + g.chars(24, base62) + id, id
	case JWT:
		return g.jwt(sessionID, carrierID, id), id
	case DBDSN:
		pass := g.chars(10, base62) + id
		return "postgres://appuser:" + pass + "@db.thu-nghiem.local:5432/app", id
	case WifiPassword, AccountPassword, ResetToken:
		return g.chars(10, base62) + id, id
	case EmailAddress:
		num := g.digits(2)
		return g.emailLocal() + num + "@thu-nghiem.vn", num
	case ThreadContent, FictionalFact:
		fact := g.fact()
		return fact, ""
	case CCCD:
		provinces := []string{"001", "002", "031", "048", "079", "092"}
		tail := g.digits(5)
		body := provinces[g.rnd.Intn(len(provinces))] + g.digits(1) + fmt.Sprintf("%02d", 70+g.rnd.Intn(36)) + tail
		v := body + checksumDigit(body)
		return v, v[6:]
	case Phone:
		prefixes := []string{"03", "05", "07", "08", "09"}
		v := prefixes[g.rnd.Intn(len(prefixes))] + g.digits(8)
		return v, v[6:]
	case BankAccount:
		v := g.digits(12)
		return v, v[6:]
	case Card:
		// ADR-0003: custom 90-prefixed range, intentionally fails Luhn.
		v := "90" + g.digits(14)
		if LuhnValid(v) {
			last := (v[15]-'0'+1)%10 + '0'
			v = v[:15] + string(last)
		}
		return v, v[10:]
	case TaxID:
		v := g.digits(10)
		return v, v[4:]
	case Address:
		streets := []string{"Hoa Lan", "Sao Mai", "Hồng Lạc", "Bình Minh", "Thuận Kiều"}
		v := fmt.Sprintf("%s %s, Phường %d, Quận %d", g.digits(2), streets[g.rnd.Intn(len(streets))], 1+g.rnd.Intn(12), 1+g.rnd.Intn(12))
		return v, ""
	default:
		return g.chars(16, base62), id
	}
}

func checksumDigit(body string) string {
	sum := 0
	for _, c := range body {
		sum += int(c - '0')
	}
	return fmt.Sprintf("%d", sum%10)
}

func (g *Generator) jwt(sessionID, carrierID, id string) string {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, _ := json.Marshal(map[string]any{"spill": 1, "sid": sessionID, "cid": carrierID, "jti": id})
	enc := base64.RawURLEncoding.EncodeToString
	signing := enc(header) + "." + enc(payload)
	key := sha256.Sum256([]byte(fmt.Sprintf("spillcheck-%d-%s", g.seed, sessionID)))
	mac := hmac.New(sha256.New, key[:])
	mac.Write([]byte(signing))
	return signing + "." + enc(mac.Sum(nil))
}

// DecodeJWTPayload decodes (without verifying) a JWT payload for tests
// and detectors.
func DecodeJWTPayload(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("bait: jwt has %d parts", len(parts))
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var out map[string]any
	return out, json.Unmarshal(raw, &out)
}

// Fingerprints returns the SHA-256 of the full value plus every
// 12-character sliding window (PLAN 7.3).
func Fingerprints(value string) []string {
	h := sha256.Sum256([]byte(value))
	out := []string{hex.EncodeToString(h[:])}
	if len(value) >= 12 {
		seen := map[string]bool{out[0]: true}
		for i := 0; i+12 <= len(value); i++ {
			wh := sha256.Sum256([]byte(value[i : i+12]))
			s := hex.EncodeToString(wh[:])
			if !seen[s] {
				seen[s] = true
				out = append(out, s)
			}
		}
	}
	return out
}

// LuhnValid reports whether digits passes the Luhn check.
func LuhnValid(digits string) bool {
	sum, alt := 0, false
	for i := len(digits) - 1; i >= 0; i-- {
		d := int(digits[i] - '0')
		if d < 0 || d > 9 {
			return false
		}
		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}
	return len(digits) > 0 && sum%10 == 0
}

var vnFamily = []string{"Nguyễn", "Trần", "Lê", "Phạm", "Hoàng", "Huỳnh", "Vũ", "Đặng", "Bùi", "Đỗ"}
var vnMiddle = []string{"Thị", "Văn", "Quốc", "Minh", "Thanh", "Ngọc", "Hồng", "Đức"}
var vnGiven = []string{"Mai", "Lan", "Hùng", "Tuấn", "Linh", "Phương", "Anh", "Khoa", "Trang", "Dũng"}

func (g *Generator) emailLocal() string {
	given := vnGiven[g.rnd.Intn(len(vnGiven))]
	family := vnFamily[g.rnd.Intn(len(vnFamily))]
	return strings.ToLower(stripAccents(given)) + "." + strings.ToLower(stripAccents(family))
}

func (g *Generator) fact() string {
	projects := []string{"Hồng Lạc", "Sao Mai", "Bình Minh Sao", "Thuận Kiều Xanh"}
	partners := []string{"Bình Minh Sao", "Hồng Lạc Phát", "Sao Mai Group"}
	return fmt.Sprintf("Dự án %s, giá trị %d,%02d tỷ, ký với công ty %s",
		projects[g.rnd.Intn(len(projects))], 1+g.rnd.Intn(9), g.rnd.Intn(100), partners[g.rnd.Intn(len(partners))])
}

func stripAccents(s string) string {
	repl := map[rune]rune{'à': 'a', 'á': 'a', 'ả': 'a', 'ã': 'a', 'ạ': 'a', 'ă': 'a', 'ắ': 'a', 'ằ': 'a', 'ặ': 'a', 'â': 'a', 'ấ': 'a', 'ầ': 'a', 'ậ': 'a', 'è': 'e', 'é': 'e', 'ẻ': 'e', 'ẹ': 'e', 'ê': 'e', 'ế': 'e', 'ề': 'e', 'ệ': 'e', 'ì': 'i', 'í': 'i', 'ị': 'i', 'ò': 'o', 'ó': 'o', 'ỏ': 'o', 'ọ': 'o', 'ô': 'o', 'ố': 'o', 'ồ': 'o', 'ộ': 'o', 'ơ': 'o', 'ớ': 'o', 'ờ': 'o', 'ợ': 'o', 'ù': 'u', 'ú': 'u', 'ủ': 'u', 'ụ': 'u', 'ư': 'u', 'ứ': 'u', 'ừ': 'u', 'ự': 'u', 'ỳ': 'y', 'ý': 'y', 'đ': 'd', 'Đ': 'D'}
	var b strings.Builder
	for _, r := range s {
		if v, ok := repl[r]; ok {
			b.WriteRune(v)
		} else if r < 128 {
			b.WriteRune(r)
		}
	}
	return b.String()
}
