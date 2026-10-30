// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package persona generates a fake Vietnamese persona per session
// (PLAN 7.4, ADR-0004). Names are random per session with an
// exclusion list; no real person's data is used.
package persona

import (
	"fmt"
	"math/rand"
	"strings"
)

// Persona is a fake identity for one session.
type Persona struct {
	SessionID       string
	FullName        string
	Email           string
	Phone           string
	Company         string
	Title           string
	WebmailPassword string
	IsDemo          bool
}

var family = []string{"Nguyễn", "Trần", "Lê", "Phạm", "Hoàng", "Huỳnh", "Vũ", "Đặng", "Bùi", "Đỗ"}
var middle = []string{"Thị", "Văn", "Quốc", "Minh", "Thanh", "Ngọc", "Hồng", "Đức"}
var given = []string{"Mai", "Lan", "Hùng", "Tuấn", "Linh", "Phương", "Anh", "Khoa", "Trang", "Dũng"}
var companies = []string{"Bình Minh Sao", "Hồng Lạc Phát", "Sao Mai Group", "Thuận Kiều Xanh"}
var titles = []string{"Kế toán", "Nhân viên văn phòng", "Trưởng nhóm", "Chủ cửa hàng"}

// ExcludedNames returns the blocklist of well-known public figures
// that a generated persona must never equal (ADR-0004).
func ExcludedNames() []string {
	return []string{"Hồ Chí Minh", "Võ Nguyên Giáp", "Nguyễn Ái Quốc"}
}

// Generate creates a random persona for a session.
func Generate(seed int64, sessionID string) (Persona, error) {
	rnd := rand.New(rand.NewSource(seed)) //nolint:gosec // fake data, determinism required for tests/demo
	for attempt := 0; attempt < 100; attempt++ {
		name := fmt.Sprintf("%s %s %s",
			family[rnd.Intn(len(family))], middle[rnd.Intn(len(middle))], given[rnd.Intn(len(given))])
		banned := false
		for _, b := range ExcludedNames() {
			if name == b {
				banned = true
			}
		}
		if banned {
			continue
		}
		local := strings.ToLower(stripAccents(strings.Split(name, " ")[2])) + "." +
			strings.ToLower(stripAccents(strings.Split(name, " ")[0]))
		return Persona{
			SessionID: sessionID, FullName: name,
			Email:           fmt.Sprintf("%s%02d@thu-nghiem.vn", local, rnd.Intn(100)),
			Phone:           fmt.Sprintf("09%08d", rnd.Intn(100000000)),
			Company:         companies[rnd.Intn(len(companies))],
			Title:           titles[rnd.Intn(len(titles))],
			WebmailPassword: fmt.Sprintf("Spill-%s-%04d", stripAccents(strings.Split(name, " ")[2]), rnd.Intn(10000)),
		}, nil
	}
	return Persona{}, fmt.Errorf("persona: could not generate a non-excluded name")
}

// Demo returns the fixed, clearly-labelled demo persona (ADR-0004).
func Demo() Persona {
	return Persona{
		FullName: "Người Dùng Thử Nghiệm", Email: "thu.nghiem@thu-nghiem.vn",
		Phone: "0900000000", Company: "Công ty Thử Nghiệm", Title: "Người dùng demo",
		WebmailPassword: "Spill-Demo-0000", IsDemo: true,
	}
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
