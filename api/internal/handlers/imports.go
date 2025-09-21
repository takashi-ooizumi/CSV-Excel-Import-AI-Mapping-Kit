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
	return func (w http.ResponseWriter, r *http.Request)  {
		// サイズ制限
		r.Body = http.MaxBytesReader(w, r.Body, int64(maxUploadMB << 20)) // 20MB

		// multipart 取得
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file is required (multipart/form-data)", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// 全体をバッファ（小〜中サイズ前提。大きくなればストリーミングに変更）
		var buf bytes.Buffer
		if _, err:= io.Copy(&buf, file); err != nil {
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
		reader.LazyQuotes = true // 厳密なクオートチェックをしない

		// 先頭行をpeek
		all := make([][]string, 0, previewRows+1)
		for i := 0; i < previewRows + 1; i++ {
			rec, err := reader.Read()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				http.Error(w, "csv parse error: " + err.Error(), http.StatusBadRequest)
				return
			}
			all = append(all, rec)
		}

		var headers []string
		hasHeader := false
		rows := all
		if len(all) > 0 && looksLikeHeader(all[0]) {
			hasHeader = true
			headers = normalizeHeaders(all[0])
			if len(all) > 1 {
				rows = all[1:]
			} else {
				rows = [][]string{}
			}
		} else if len(all) > 0 {
			// ヘッダーなしならカラム名を自動生成
			headers = make([]string, len(all[0]))
			for i := range headers {
				headers[i] = "col_" + strconv.Itoa(i+1)
			}
		}

		// レスポンス
		resp := previewResponse{
			Delimiter: del,
			HasHeader: hasHeader,
			Headers: headers,
			SampleRows: rows,
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

func looksLikeHeader(rec []string) bool {
	// 漢字　数字　アンダースコア　スペース　ハイフン程度で構成、かつ重複と空が少ないならヘッダーと推測
	if len(rec) == 0 {
		return false
	}
	seen := map[string]struct{}{}
	empties := 0
	for _, v := range rec {
		name := strings.TrimSpace(v)
		if name == "" {
			empties++
			continue
		}
		// 記号だらけならデータ行の可能性
		valid := 0
		for _, r := range name {
			if r == '_' || r == '-' || r == ' ' || r == '/' || r == '.' {
				valid++
				continue
			}
			if r >= '0' && r <= '9' {
				// 数値だらけの列名は微妙
				continue
			}
			if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r > 127{
				valid++
			}
		}
		if valid == 0 {
			return false
		}
		if _, ok := seen[strings.ToLower(name)]; ok {
			// 重複カラム名かも、ヘッダー性を下げる
			return false
		}
		seen[strings.ToLower(name)] = struct{}{}
	}
	return empties <= len(rec) / 3
}

func normalizeHeaders(rec []string) []string {
	out := make([]string, len(rec))
	used := map[string]int{}
	for i, v := range rec {
		name := strings.TrimSpace(v)
		if name == "" {
			name = "col_" + strconv.Itoa(i+1)
		}
		// シンプルに正規化
		name = strings.ToLower(name)
		name = strings.ReplaceAll(name, " ", "_")
		name = strings.ReplaceAll(name, "-", "_")
		// 非ASCIIはそのまま許容（可視性だけ注意）
		if !utf8.ValidString(name) {
			name = "col_" + strconv.Itoa(i+1)
		}
		// 重複回避
		key := name
		if _, ok := used[key]; ok {
			used[key] = 1
		} else {
			cnt := used[key]
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


