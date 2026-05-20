package reader

import (
	"archive/tar"
	"log-compare/config"
	"os"
	"path/filepath"
	"testing"
)

// --- detectDelimiter ---

func TestDetectDelimiter_PipePlusPipe(t *testing.T) {
	d, err := detectDelimiter("a|++|b|++|c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != "|++|" {
		t.Errorf("got %q, want |++|", d)
	}
}

func TestDetectDelimiter_Comma(t *testing.T) {
	d, err := detectDelimiter("a,b,c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != "," {
		t.Errorf("got %q, want ,", d)
	}
}

func TestDetectDelimiter_Tab(t *testing.T) {
	d, err := detectDelimiter("a\tb\tc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != "\t" {
		t.Errorf("got %q, want \\t", d)
	}
}

func TestDetectDelimiter_Pipe(t *testing.T) {
	d, err := detectDelimiter("a|b|c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != "|" {
		t.Errorf("got %q, want |", d)
	}
}

func TestDetectDelimiter_NoDelimiter(t *testing.T) {
	_, err := detectDelimiter("abc")
	if err == nil {
		t.Error("expected error for unrecognizable delimiter")
	}
}

// --- splitLine ---

func TestSplitLine_Basic(t *testing.T) {
	fields := splitLine("a|++|b|++|c", "|++|")
	if len(fields) != 3 {
		t.Fatalf("got %d fields, want 3", len(fields))
	}
	if fields[0] != "a" || fields[1] != "b" || fields[2] != "c" {
		t.Errorf("fields = %v", fields)
	}
}

func TestSplitLine_TrimSpace(t *testing.T) {
	fields := splitLine(" a , b , c ", ",")
	if fields[0] != "a" || fields[1] != "b" || fields[2] != "c" {
		t.Errorf("fields not trimmed: %v", fields)
	}
}

func TestSplitLine_SingleField(t *testing.T) {
	fields := splitLine("only", ",")
	if len(fields) != 1 || fields[0] != "only" {
		t.Errorf("got %v", fields)
	}
}

// --- isTarFile ---

func TestIsTarFile(t *testing.T) {
	if !isTarFile("foo.tar") {
		t.Error("expected true for .tar")
	}
	if isTarFile("foo.txt") {
		t.Error("expected false for .txt")
	}
	if isTarFile("foo.tar.gz") {
		t.Error("expected false for .tar.gz")
	}
}

// --- findFiles ---

func TestFindFiles_MatchPattern(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CMD_001.txt"), []byte("header\n"), 0644)
	os.WriteFile(filepath.Join(dir, "DES_001.txt"), []byte("header\n"), 0644)
	os.WriteFile(filepath.Join(dir, "other.log"), []byte("data\n"), 0644)

	files, err := findFiles(dir, "CMD_*.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}
	if filepath.Base(files[0]) != "CMD_001.txt" {
		t.Errorf("got %s", filepath.Base(files[0]))
	}
}

func TestFindFiles_Recursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(sub, "b.txt"), []byte("y"), 0644)

	files, err := findFiles(dir, "*.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("got %d files, want 2", len(files))
	}
}

func TestFindFiles_NoMatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0644)

	files, err := findFiles(dir, "CMD_*.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("got %d files, want 0", len(files))
	}
}

func TestFindFiles_NonExistentDir(t *testing.T) {
	files, err := findFiles("/nonexistent_dir_xyz", "*.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("got %d files, want 0", len(files))
	}
}

// --- readTextFile ---

func TestReadTextFile_WithDelimiter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "h1,h2,h3\nv1,v2,v3\nv4,v5,v6\n"
	os.WriteFile(path, []byte(content), 0644)

	ltCfg := &config.LogTypeConfig{Delimiter: ","}
	result, err := readTextFile(path, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Header) != 3 || result.Header[0] != "h1" {
		t.Errorf("header = %v", result.Header)
	}
	if len(result.Records) != 2 {
		t.Fatalf("got %d records, want 2", len(result.Records))
	}
	if result.Records[0].Fields[0] != "v1" {
		t.Errorf("record[0] = %v", result.Records[0].Fields)
	}
	if result.Records[0].LineNumber != 2 {
		t.Errorf("line number = %d, want 2", result.Records[0].LineNumber)
	}
	if result.Records[0].FilePath != path {
		t.Errorf("file path = %s", result.Records[0].FilePath)
	}
}

func TestReadTextFile_AutoDetectDelimiter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "h1|++|h2\nv1|++|v2\n"
	os.WriteFile(path, []byte(content), 0644)

	ltCfg := &config.LogTypeConfig{} // no delimiter set
	result, err := readTextFile(path, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Records) != 1 {
		t.Fatalf("got %d records, want 1", len(result.Records))
	}
	if result.Records[0].Fields[0] != "v1" || result.Records[0].Fields[1] != "v2" {
		t.Errorf("fields = %v", result.Records[0].Fields)
	}
}

func TestReadTextFile_SkipsEmptyLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "h1,h2\n\nv1,v2\n  \nv3,v4\n"
	os.WriteFile(path, []byte(content), 0644)

	ltCfg := &config.LogTypeConfig{Delimiter: ","}
	result, err := readTextFile(path, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Records) != 2 {
		t.Errorf("got %d records, want 2", len(result.Records))
	}
}

func TestReadTextFile_NoDataLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("h1,h2\n"), 0644)

	ltCfg := &config.LogTypeConfig{Delimiter: ","}
	result, err := readTextFile(path, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Records) != 0 {
		t.Errorf("got %d records, want 0", len(result.Records))
	}
	if len(result.Header) != 2 {
		t.Errorf("header = %v", result.Header)
	}
}

func TestReadTextFile_FileNotFound(t *testing.T) {
	ltCfg := &config.LogTypeConfig{Delimiter: ","}
	_, err := readTextFile("/nonexistent/file.txt", ltCfg)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestReadTextFile_BadDelimiter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("no_delimiter_here\n"), 0644)

	ltCfg := &config.LogTypeConfig{} // no delimiter, auto-detect will fail
	_, err := readTextFile(path, ltCfg)
	if err == nil {
		t.Error("expected error for unrecognizable delimiter")
	}
}

// --- readTarFile ---

func createTestTar(t *testing.T, dir, tarName string, entries map[string]string) string {
	t.Helper()
	tarPath := filepath.Join(dir, tarName)
	f, err := os.Create(tarPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	for name, content := range entries {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	return tarPath
}

func TestReadTarFile_Basic(t *testing.T) {
	dir := t.TempDir()
	tarPath := createTestTar(t, dir, "test.tar", map[string]string{
		"log.txt": "h1,h2,h3\nv1,v2,v3\nv4,v5,v6\n",
	})

	ltCfg := &config.LogTypeConfig{Delimiter: ","}
	result, err := readTarFile(tarPath, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Header) != 3 || result.Header[0] != "h1" {
		t.Errorf("header = %v", result.Header)
	}
	if len(result.Records) != 2 {
		t.Fatalf("got %d records, want 2", len(result.Records))
	}
	if result.Records[0].Fields[0] != "v1" {
		t.Errorf("fields = %v", result.Records[0].Fields)
	}
	if result.Records[0].TarName != "log.txt" {
		t.Errorf("tar name = %s", result.Records[0].TarName)
	}
	if result.Records[0].LineNumber != 2 {
		t.Errorf("line number = %d", result.Records[0].LineNumber)
	}
}

func TestReadTarFile_SkipsNonTxt(t *testing.T) {
	dir := t.TempDir()
	tarPath := createTestTar(t, dir, "test.tar", map[string]string{
		"data.csv":  "a,b\n1,2\n",
		"log.txt":   "h1\nv1\n",
		"readme.md": "# title\n",
	})

	ltCfg := &config.LogTypeConfig{Delimiter: ","}
	result, err := readTarFile(tarPath, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Records) != 1 {
		t.Errorf("got %d records, want 1", len(result.Records))
	}
}

func TestReadTarFile_MultipleTxtEntries(t *testing.T) {
	dir := t.TempDir()
	tarPath := createTestTar(t, dir, "test.tar", map[string]string{
		"a.txt": "h1\nv1\n",
		"b.txt": "h1\nv2\nv3\n",
	})

	ltCfg := &config.LogTypeConfig{Delimiter: ","}
	result, err := readTarFile(tarPath, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1 from a.txt + 2 from b.txt = 3
	if len(result.Records) != 3 {
		t.Errorf("got %d records, want 3", len(result.Records))
	}
}

func TestReadTarFile_AutoDetectDelimiter(t *testing.T) {
	dir := t.TempDir()
	tarPath := createTestTar(t, dir, "test.tar", map[string]string{
		"log.txt": "h1|++|h2\nv1|++|v2\n",
	})

	ltCfg := &config.LogTypeConfig{} // no delimiter
	result, err := readTarFile(tarPath, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Records[0].Fields[0] != "v1" || result.Records[0].Fields[1] != "v2" {
		t.Errorf("fields = %v", result.Records[0].Fields)
	}
}

func TestReadTarFile_BadDelimiter(t *testing.T) {
	dir := t.TempDir()
	tarPath := createTestTar(t, dir, "test.tar", map[string]string{
		"log.txt": "no_delimiter\n",
	})

	ltCfg := &config.LogTypeConfig{}
	_, err := readTarFile(tarPath, ltCfg)
	if err == nil {
		t.Error("expected error for unrecognizable delimiter")
	}
}

func TestReadTarFile_FileNotFound(t *testing.T) {
	ltCfg := &config.LogTypeConfig{Delimiter: ","}
	_, err := readTarFile("/nonexistent/file.tar", ltCfg)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

// --- ReadLogs (integration) ---

func TestReadLogs_TextFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CMD_001.txt"), []byte("h1,h2\na,b\nc,d\n"), 0644)
	os.WriteFile(filepath.Join(dir, "CMD_002.txt"), []byte("h1,h2\ne,f\n"), 0644)
	os.WriteFile(filepath.Join(dir, "DES_001.txt"), []byte("h1,h2\nx,y\n"), 0644)

	ltCfg := &config.LogTypeConfig{
		Delimiter:   ",",
		FilePattern: "CMD_*.txt",
	}
	result, err := ReadLogs(dir, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Records) != 3 {
		t.Errorf("got %d records, want 3", len(result.Records))
	}
	if len(result.Header) != 2 {
		t.Errorf("header = %v", result.Header)
	}
}

func TestReadLogs_TarFiles(t *testing.T) {
	dir := t.TempDir()
	tarPath := filepath.Join(dir, "CMD_test.tar")
	f, _ := os.Create(tarPath)
	tw := tar.NewWriter(f)
	content := "h1,h2\na,b\nc,d\n"
	hdr := &tar.Header{Name: "log.txt", Mode: 0644, Size: int64(len(content))}
	tw.WriteHeader(hdr)
	tw.Write([]byte(content))
	tw.Close()
	f.Close()

	ltCfg := &config.LogTypeConfig{
		Delimiter:   ",",
		FilePattern: "CMD_*.tar",
	}
	result, err := ReadLogs(dir, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Records) != 2 {
		t.Errorf("got %d records, want 2", len(result.Records))
	}
}

func TestReadLogs_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	ltCfg := &config.LogTypeConfig{
		Delimiter:   ",",
		FilePattern: "CMD_*.txt",
	}
	result, err := ReadLogs(dir, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Records) != 0 {
		t.Errorf("got %d records, want 0", len(result.Records))
	}
}

func TestReadLogs_SkippedFiles(t *testing.T) {
	dir := t.TempDir()
	// valid file with pipe delimiter
	os.WriteFile(filepath.Join(dir, "CMD_001.txt"), []byte("h1|h2\nv1|v2\n"), 0644)
	// file with no recognizable delimiter (will be skipped)
	os.WriteFile(filepath.Join(dir, "CMD_002.txt"), []byte("no_delimiter_here\n"), 0644)

	ltCfg := &config.LogTypeConfig{
		FilePattern: "CMD_*.txt",
	}
	result, err := ReadLogs(dir, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Records) != 1 {
		t.Errorf("got %d records, want 1", len(result.Records))
	}
	if len(result.SkippedFiles) != 1 {
		t.Errorf("got %d skipped files, want 1", len(result.SkippedFiles))
	}
}

func TestReadLogs_MixedTxtAndTar(t *testing.T) {
	dir := t.TempDir()
	// txt file
	os.WriteFile(filepath.Join(dir, "CMD_001.txt"), []byte("h1,h2\na,b\n"), 0644)
	// tar file
	tarPath := filepath.Join(dir, "CMD_002.tar")
	f, _ := os.Create(tarPath)
	tw := tar.NewWriter(f)
	content := "h1,h2\nc,d\n"
	thdr := &tar.Header{Name: "log.txt", Mode: 0644, Size: int64(len(content))}
	tw.WriteHeader(thdr)
	tw.Write([]byte(content))
	tw.Close()
	f.Close()

	ltCfg := &config.LogTypeConfig{
		Delimiter:   ",",
		FilePattern: "CMD_00*",
	}
	result, err := ReadLogs(dir, ltCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Records) != 2 {
		t.Errorf("got %d records, want 2", len(result.Records))
	}
}
