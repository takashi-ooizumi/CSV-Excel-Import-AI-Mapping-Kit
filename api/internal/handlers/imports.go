package handlers

import (
	//"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"

	//"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

const maxUploadMB = 20
const previewRows = 20

type previewResponse struct {
	Delimiter    string     `json:"delimiter"`
	HasHeader    bool       `json:"hasHeader"`
	Headers      []string   `json:"headers"`
	SampleRows   [][]string `json:"sampleRows"`
	CountGuessed int        `json:"countGuessed"`
}

func HandleUploadPreview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// サイズ制限
		r.Body = http.MaxBytesReader(w, r.Body, int64(maxUploadMB<<20)) // 20MB

		// multipart 取得
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file is required (multipart/form-data)", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// 全体をバッファ（小〜中サイズ前提。大きくなればストリーミングに変更）
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, file); err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}

		// 文字コード/BOM簡易処理（UTF-8 BOM除去）
		b := buf.Bytes()
		if bytes.HasPrefix(b, []byte{0xEF, 0xBB, 0xBF}) {
			b = b[3:]
		}

		// 区切り文字推定
		del := guessDelimiter(b)

		// CSVパーサ
		reader := csv.NewReader(bytes.NewReader(b))
		reader.Comma = rune(del[0])
		reader.FieldsPerRecord = -1 // 可変長対応
		reader.LazyQuotes = true    // 厳密なクオートチェックをしない

		// 先頭行をpeek
		all := make([][]string, 0, previewRows+1)
		for i := 0; i < previewRows+1; i++ {
			rec, err := reader.Read()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				http.Error(w, "csv parse error: "+err.Error(), http.StatusBadRequest)
				return
			}
			all = append(all, rec)
		}

		var headers []string
		hasHeader := false
		rows := all
		// ヘッダー判定
		if len(all) > 0 {
			var next []string
			if len(all) > 1 {
				next = all[1]
			}
			if looksLikeHeader(all[0], next) {
				hasHeader = true
				headers = normalizeHeaders(all[0])
				rows = all[1:]
			} else {
				// ヘッダーなしならカラム名を自動生成
				headers = make([]string, len(all[0]))
				for i := range headers {
					headers[i] = "col_" + strconv.Itoa(i+1)
				}
			}
		}

		// レスポンス
		resp := previewResponse{
			Delimiter:    del,
			HasHeader:    hasHeader,
			Headers:      headers,
			SampleRows:   rows,
			CountGuessed: len(rows),
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(resp)
	}
}

// Utility functions

func guessDelimiter(b []byte) string {
	// 先頭1KBから候補文字の出現数を数える
	candidates := []byte{',', '\t', ';', '|'}
	end := len(b)
	if end > 1024 {
		end = 1024
	}
	snipet := b[:end]
	counts := make(map[byte]int)
	for _, c := range candidates {
		counts[c] = bytes.Count(snipet, []byte{c})
	}
	best := byte(',')
	bestScore := -1
	for c, n := range counts {
		if n > bestScore {
			best = c
			bestScore = n
		}
	}
	return string([]byte{best})
}

// 先頭行がヘッダらしいか？ 2行目(next)を見てコントラスト判定も行う
func looksLikeHeader(first []string, next []string) bool {
	if len(first) == 0 {
		return false
	}
	// 指標を取る
	statsFirst := rowStats(first)
	// ヘッダの素点: 英字主体が6割以上、かつ kv/numeric/datetime が多すぎない
	headerish := statsFirst.alphaWordRatio >= 0.60 &&
		statsFirst.keyValueRatio < 0.20 &&
		statsFirst.numericLikeRatio < 0.40 &&
		statsFirst.datetimeLikeRatio < 0.40 &&
		statsFirst.emptyRatio <= 0.34 &&
		!statsFirst.hasDup

	if !headerish {
		return false
	}
	// 2行目があるなら、データ行っぽさ（数値・日時・kvが増える）で後押し
	if len(next) > 0 {
		statsNext := rowStats(next)
		// 次行は英字主体が少なめ、かつ数値/日時/kvが一定以上
		contrastOK := statsNext.alphaWordRatio <= 0.50 ||
			statsNext.numericLikeRatio >= 0.30 ||
			statsNext.datetimeLikeRatio >= 0.20 ||
			statsNext.keyValueRatio >= 0.20
		if !contrastOK {
			// コントラストが無いならヘッダとは言い切らない
			return false
		}
	}
	return true
}

type rowStat struct {
	alphaWordRatio    float64
	numericLikeRatio  float64
	datetimeLikeRatio float64
	keyValueRatio     float64
	emptyRatio        float64
	hasDup            bool
}

func rowStats(cols []string) rowStat {
	n := float64(len(cols))
	if n == 0 {
		return rowStat{}
	}
	seen := map[string]struct{}{}
	dup := false
	alpha, numeric, dt, kv, empty := 0, 0, 0, 0, 0

	for _, raw := range cols {
		s := strings.TrimSpace(raw)
		if s == "" {
			empty++
			continue
		}
		low := strings.ToLower(s)
		if _, ok := seen[low]; ok {
			dup = true
		} else {
			seen[low] = struct{}{}
		}

		if isKeyValue(s) {
			kv++
		}
		if isDateTimeLike(s) {
			dt++
		}
		if isNumericLike(s) {
			numeric++
		}
		if isAlphaWord(s) {
			alpha++
		}
	}

	return rowStat{
		alphaWordRatio:    float64(alpha) / n,
		numericLikeRatio:  float64(numeric) / n,
		datetimeLikeRatio: float64(dt) / n,
		keyValueRatio:     float64(kv) / n,
		emptyRatio:        float64(empty) / n,
		hasDup:            dup,
	}
}

func isAlphaWord(s string) bool {
	// 英字・空白・アンダースコアのみ（数字/記号/=.:- を含んだら除外）
	if s == "" {
		return false
	}
	for _, r := range s {
		if r == ' ' || r == '_' {
			continue
		}
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			continue
		}
		// 数字やよくある記号が混ざっていたら英字主体とはみなさない
		return false
	}
	return true
}

func isNumericLike(s string) bool {
	// 数字・小数・符号・桁区切り・百分率程度
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	dots := 0
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r == '.' || r == ',':
			dots++
			if dots > 3 {
				return false
			}
		case r == '+' || r == '-' || r == '%':
		default:
			return false
		}
	}
	return true
}

func isDateTimeLike(s string) bool {
	// ざっくり："YYYY-MM-DD" や "YYYY/MM/DD hh:mm:ss" など数字と -/: と空白で構成
	if s == "" {
		return false
	}
	hasDigit := false
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '-' || r == '/' || r == ':' || r == ' ' || r == 'T' || r == 'Z' {
			if r >= '0' && r <= '9' {
				hasDigit = true
			}
			continue
		}
		return false
	}
	return hasDigit
}

func isKeyValue(s string) bool {
	// user=1 / ip=203.0.113.10 のような key=value を判定
	if strings.Count(s, "=") != 1 {
		return false
	}
	parts := strings.SplitN(s, "=", 2)
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	if key == "" || val == "" {
		return false
	}
	// キーは英字/数字/アンダースコアのみ
	for _, r := range key {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}

func normalizeHeaders(rec []string) []string {
	out := make([]string, len(rec))
	used := map[string]int{}
	for i, v := range rec {
		name := strings.TrimSpace(v)
		if name == "" {
			name = "col_" + strconv.Itoa(i+1)
		}
		// 正規化
		name = strings.ToLower(name)
		name = strings.ReplaceAll(name, " ", "_")
		name = strings.ReplaceAll(name, "-", "_")
		if !utf8.ValidString(name) {
			name = "col_" + strconv.Itoa(i+1)
		}

		// 重複回避（修正後）
		key := name
		if _, ok := used[key]; !ok {
			used[key] = 1 // 初回：そのまま
		} else {
			cnt := used[key] // 2回目以降：_1, _2...
			used[key] = cnt + 1
			name = key + "_" + strconv.Itoa(cnt)
		}

		out[i] = name
	}
	return out
}

// CORS Middleware（複数Originをカンマ区切りで）
func CORSMiddleware() func(next http.Handler) http.Handler {
	// 両方の名前をサポート（複数形/単数形）
	raw := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("ALLOWED_ORIGIN"))
	}
	// 開発用デフォルト（未設定なら localhost:3000 を許可）
	if raw == "" {
		raw = "http://localhost:3000"
	}

	origins := strings.Split(raw, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := ""
			for _, o := range origins {
				if o == "*" {
					allowed = "*"
					break
				}
				if o != "" && origin == o {
					allowed = o
					break
				}
			}

			// 許可が決まったときだけヘッダを付与
			if allowed != "" {
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Origin", allowed)
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			}

			// プリフライトは許可が決まった時だけ 204 を返す
			if r.Method == http.MethodOptions && allowed != "" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
