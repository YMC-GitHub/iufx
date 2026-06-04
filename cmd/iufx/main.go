package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

// 命令常量（避免拼写错误）
const (
	CmdAnno       = "anno"
	CmdLabel      = "label"
	CmdPick       = "pick"
	CmdLabelCheck = "label/check"

	Version = "v7.4.0"
	Usage   = `
iufx - Annotation Tool for Bamboo Datasets

Usage:
  iufx --cmd <command> [options]
  iufx help
  iufx version

Commands:
  anno                Sort CSV by filename and x/y
  label               Convert CSV to YOLO-style label files
  pick                Copy images listed in CSV to output folder
  label/check         Visualize labels on images

Global Options:
  --help              Show this help message
  --version           Show version information

Command Options:
  --cmd <command>     Command to execute (anno|label|pick|label/check)

Anno Options:
  --csv FILE          Input CSV
  --out-csv FILE      Output sorted CSV
  --dedup             Enable duplicate removal (default: true)
  --no-dedup          Disable duplicate removal

Label Options:
  --shape rect/circle Shape type (default: circle)
  --box-size N        Box size (default: 88)
  --input FILE        CSV input
  --output DIR        Label output directory

Pick Options:
  --input FILE        CSV input
  --output DIR        Image output directory
  --image-source DIR  Source image folder

Label/Check Options:
  --image-source DIR  Images folder
  --check-output DIR  Output checked images
  --output DIR        Labels folder
  --check-mode box    Check mode
  --shape rect/circle Shape (default: circle)
  --box-size N        Box size (default: 88)
  --shape-color COLOR Shape color (default: yellow)
  --shape-size N      Line width (default: 10)
  --text-color COLOR  Text color (default: red)
  --text              Enable index text
  --sort              Sort points before drawing
  --text-size N       Font size (default: 78)
`
)

// 全局字体（只加载一次）
var globalFont font.Face

func init() {
	f, _ := truetype.Parse(goregular.TTF)
	globalFont = truetype.NewFace(f, &truetype.Options{Size: 78})
}

// ==============================================================================================
// 数据结构定义
// ==============================================================================================

type LabelPoint struct {
	X, Y int
}

type AnnotationPoint struct {
	Row      []string
	X, Y     int
	Width    int
	Height   int
	Filename string
	Class    string
}

type YOLOLabel struct {
	ClassID int
	CenterX float64
	CenterY float64
	Width   float64
	Height  float64
}

// CheckConfig 检查命令配置结构体（替代大量flag传递）
type CheckConfig struct {
	ImageSource string
	LabelDir    string
	OutputDir   string
	CheckMode   string
	Shape       string
	ShapeColor  color.Color
	TextColor   color.Color
	ShapeSize   int
	TextSize    int
	BoxSize     int
	DrawText    bool
	SortPoints  bool
}

// ==============================================================================================
// 纯函数：数据解析和验证
// ==============================================================================================

// ParseCSV 流式读取CSV（替代ReadAll，低内存占用）
func ParseCSV(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var rows [][]string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv failed: %w", err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// ValidateAnnotationRow 验证标注行是否有效
func ValidateAnnotationRow(row []string) bool {
	return len(row) >= 6
}

// ParseAnnotationRow 解析单行标注数据 + 坐标合法性校验
func ParseAnnotationRow(row []string) (*AnnotationPoint, error) {
	if !ValidateAnnotationRow(row) {
		return nil, fmt.Errorf("invalid row: need at least 6 columns, got %d", len(row))
	}

	x, err := strconv.Atoi(row[1])
	if err != nil {
		return nil, fmt.Errorf("invalid x coordinate: %w", err)
	}

	y, err := strconv.Atoi(row[2])
	if err != nil {
		return nil, fmt.Errorf("invalid y coordinate: %w", err)
	}

	width, err := strconv.Atoi(row[4])
	if err != nil {
		return nil, fmt.Errorf("invalid width: %w", err)
	}

	height, err := strconv.Atoi(row[5])
	if err != nil {
		return nil, fmt.Errorf("invalid height: %w", err)
	}

	// 坐标合法性校验：禁止负数
	if x < 0 || y < 0 || width <= 0 || height <= 0 {
		return nil, fmt.Errorf("coordinates and size must be positive")
	}

	return &AnnotationPoint{
		Row:      row,
		X:        x,
		Y:        y,
		Width:    width,
		Height:   height,
		Filename: row[3],
		Class:    row[0],
	}, nil
}

// ParseAllAnnotations 解析所有标注行
func ParseAllAnnotations(rows [][]string) ([]AnnotationPoint, error) {
	var points []AnnotationPoint
	for i, row := range rows {
		point, err := ParseAnnotationRow(row)
		if err != nil {
			return nil, fmt.Errorf("parse row %d failed: %w", i, err)
		}
		points = append(points, *point)
	}
	return points, nil
}

// ==============================================================================================
// 纯函数：数据处理和转换
// ==============================================================================================

// GroupByFilename 按文件名分组
func GroupByFilename(points []AnnotationPoint) map[string][]AnnotationPoint {
	groups := make(map[string][]AnnotationPoint)
	for _, point := range points {
		groups[point.Filename] = append(groups[point.Filename], point)
	}
	return groups
}

// DeduplicatePoints 去除重复点（相同坐标）
func DeduplicatePoints(points []AnnotationPoint) []AnnotationPoint {
	seen := make(map[string]bool)
	var unique []AnnotationPoint

	for _, p := range points {
		key := fmt.Sprintf("%d,%d", p.X, p.Y)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, p)
		}
	}
	return unique
}

// SortPointsByYX 按 Y 升序，Y 相同时按 X 升序排序
func SortPointsByYX(points []AnnotationPoint) []AnnotationPoint {
	sorted := make([]AnnotationPoint, len(points))
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Y != sorted[j].Y {
			return sorted[i].Y < sorted[j].Y
		}
		return sorted[i].X < sorted[j].X
	})
	return sorted
}

// SortFilenames 对文件名进行排序
func SortFilenames(filenames []string) []string {
	sorted := make([]string, len(filenames))
	copy(sorted, filenames)
	sort.Strings(sorted)
	return sorted
}

// ExtractUniqueFilenames 提取唯一的文件名列表
func ExtractUniqueFilenames(points []AnnotationPoint) []string {
	filenameMap := make(map[string]bool)
	for _, p := range points {
		filenameMap[p.Filename] = true
	}

	filenames := make([]string, 0, len(filenameMap))
	for f := range filenameMap {
		filenames = append(filenames, f)
	}
	return SortFilenames(filenames)
}

// ==============================================================================================
// 纯函数：YOLO 格式转换
// ==============================================================================================

// ConvertToYOLO 将标注点转换为 YOLO 格式
func ConvertToYOLO(point AnnotationPoint, boxSize int) YOLOLabel {
	cx := float64(point.X) / float64(point.Width)
	cy := float64(point.Y) / float64(point.Height)
	size := float64(boxSize) / float64(point.Width)

	return YOLOLabel{
		ClassID: 0,
		CenterX: cx,
		CenterY: cy,
		Width:   size,
		Height:  size,
	}
}

// FormatYOLOLine 格式化 YOLO 标签为字符串
func FormatYOLOLine(label YOLOLabel) string {
	return fmt.Sprintf("%d %.6f %.6f %.6f %.6f",
		label.ClassID, label.CenterX, label.CenterY, label.Width, label.Height)
}

// ==============================================================================================
// 纯函数：CSV 写入
// ==============================================================================================

// WriteCSVRows 将数据行写入 CSV 文件
func WriteCSVRows(filePath string, rows [][]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write row failed: %w", err)
		}
	}
	return nil
}

// ==============================================================================================
// 纯函数：文件操作
// ==============================================================================================

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file failed: %w", err)
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination file failed: %w", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("copy file failed: %w", err)
	}
	return nil
}

// EnsureDir 确保目录存在
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// IsImageFile 校验是否为图片文件
func IsImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

// ==============================================================================================
// 纯函数：颜色处理
// ==============================================================================================

// ParseColor 解析颜色字符串
func ParseColor(s string) color.Color {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "red":
		return color.RGBA{255, 0, 0, 255}
	case "yellow":
		return color.RGBA{255, 255, 0, 255}
	case "blue":
		return color.RGBA{0, 0, 255, 255}
	case "green":
		return color.RGBA{0, 255, 0, 255}
	case "white":
		return color.RGBA{255, 255, 255, 255}
	}
	if strings.HasPrefix(s, "rgba(") && strings.HasSuffix(s, ")") {
		inner := strings.TrimPrefix(s, "rgba(")
		inner = strings.TrimSuffix(inner, ")")
		parts := strings.Split(inner, ",")
		if len(parts) == 4 {
			r := parseUint8(parts[0])
			g := parseUint8(parts[1])
			b := parseUint8(parts[2])
			a := parseUint8(parts[3])
			return color.RGBA{r, g, b, a}
		}
	}
	return color.RGBA{255, 0, 0, 255}
}

// parseUint8 解析 uint8 值
func parseUint8(s string) uint8 {
	s = strings.TrimSpace(s)
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// ==============================================================================================
// 业务逻辑函数
// ==============================================================================================

// ProcessAnnotationCSV 处理标注 CSV 的核心逻辑
func ProcessAnnotationCSV(rows [][]string, dedup bool) (map[string][]AnnotationPoint, int, error) {
	points, err := ParseAllAnnotations(rows)
	if err != nil {
		return nil, 0, err
	}

	groups := GroupByFilename(points)

	duplicateCount := 0
	for filename, group := range groups {
		originalCount := len(group)

		if dedup {
			group = DeduplicatePoints(group)
			duplicateCount += originalCount - len(group)
		}

		group = SortPointsByYX(group)
		groups[filename] = group
	}

	return groups, duplicateCount, nil
}

// GenerateYOLOLabels 生成 YOLO 标签文件
func GenerateYOLOLabels(points []AnnotationPoint, boxSize int, outputDir string) error {
	if err := EnsureDir(outputDir); err != nil {
		return err
	}

	groups := GroupByFilename(points)
	filenames := extractFilenamesFromGroups(groups)

	for _, filename := range filenames {
		group := groups[filename]
		baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
		outputPath := filepath.Join(outputDir, baseName+".txt")

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("create label file failed: %w", err)
		}

		sortedGroup := SortPointsByYX(group)

		for _, point := range sortedGroup {
			yoloLabel := ConvertToYOLO(point, boxSize)
			line := FormatYOLOLine(yoloLabel)
			fmt.Fprintln(file, line)
		}
		file.Close()

		fmt.Printf("   %s.txt (%d points)\n", baseName, len(group))
	}

	return nil
}

// ==============================================================================================
// 命令处理函数
// ==============================================================================================

// RunAnnoCommand 执行 anno 命令
func RunAnnoCommand(csvPath, outPath string, dedup bool) error {
	rows, err := ParseCSV(csvPath)
	if err != nil {
		return err
	}

	groups, duplicateCount, err := ProcessAnnotationCSV(rows, dedup)
	if err != nil {
		return err
	}

	var outputRows [][]string
	filenames := extractFilenamesFromGroups(groups)

	for _, filename := range filenames {
		for _, point := range groups[filename] {
			outputRows = append(outputRows, point.Row)
		}
	}

	if err := WriteCSVRows(outPath, outputRows); err != nil {
		return err
	}

	printStats(filenames, groups, duplicateCount, dedup)
	return nil
}

// extractFilenamesFromGroups 从分组中提取文件名列表
func extractFilenamesFromGroups(groups map[string][]AnnotationPoint) []string {
	filenames := make([]string, 0, len(groups))
	for f := range groups {
		filenames = append(filenames, f)
	}
	return SortFilenames(filenames)
}

// printStats 打印统计信息
func printStats(filenames []string, groups map[string][]AnnotationPoint, duplicateCount int, dedup bool) {
	if dedup && duplicateCount > 0 {
		fmt.Printf("✅ Sort completed (removed %d duplicate points)\n", duplicateCount)
	} else if dedup {
		fmt.Println("✅ Sort completed (no duplicate points)")
	} else {
		fmt.Println("✅ Sort completed (all points retained)")
	}

	for _, f := range filenames {
		fmt.Printf("   %s (%d points)\n", f, len(groups[f]))
	}
}

// RunLabelCommand 执行 label 命令
func RunLabelCommand(inputPath, outputDir, shape string, boxSize int) error {
	rows, err := ParseCSV(inputPath)
	if err != nil {
		return err
	}

	points, err := ParseAllAnnotations(rows)
	if err != nil {
		return err
	}

	fmt.Println("✅ Label generation completed:")
	return GenerateYOLOLabels(points, boxSize, outputDir)
}

// RunPickCommand 执行 pick 命令（增加图片格式校验）
func RunPickCommand(inputPath, imageSource, outputDir string) error {
	rows, err := ParseCSV(inputPath)
	if err != nil {
		return err
	}

	points, err := ParseAllAnnotations(rows)
	if err != nil {
		return err
	}

	uniqueFilenames := ExtractUniqueFilenames(points)
	if err := EnsureDir(outputDir); err != nil {
		return err
	}

	fmt.Println("✅ Image copy completed:")
	for _, filename := range uniqueFilenames {
		// 图片格式校验
		if !IsImageFile(filename) {
			fmt.Printf("   %s (not an image, skipped)\n", filename)
			continue
		}

		src := filepath.Join(imageSource, filename)
		dst := filepath.Join(outputDir, filename)

		if err := CopyFile(src, dst); err != nil {
			fmt.Printf("   %s (not found or copy failed)\n", filename)
			continue
		}
		fmt.Printf("   %s\n", filename)
	}

	return nil
}

// RunCheckCommand 执行 label/check 命令（修复目录错误+结构体传参）
func RunCheckCommand(cfg CheckConfig) {
	if err := EnsureDir(cfg.OutputDir); err != nil {
		fmt.Printf("❌ Failed to create output directory: %v\n", err)
		return
	}

	// 加载图片文件（自动过滤非图片）
	var files []string
	filepath.Walk(cfg.ImageSource, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if IsImageFile(path) {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)

	fmt.Println("✅ Label visualization completed:")

	for _, f := range files {
		base := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
		lbl := filepath.Join(cfg.LabelDir, base+".txt")

		img, err := gg.LoadImage(f)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to load image %s: %v\n", f, err)
			continue
		}

		points := loadLabelPoints(lbl, img)
		if cfg.SortPoints {
			points = sortLabelPoints(points)
		}

		dc := gg.NewContextForImage(img)
		drawAnnotationsOnDC(dc, points, cfg.Shape, cfg.ShapeColor, cfg.TextColor,
			cfg.ShapeSize, cfg.TextSize, cfg.BoxSize, cfg.DrawText)

		outputPath := filepath.Join(cfg.OutputDir, filepath.Base(f))
		if err := dc.SavePNG(outputPath); err != nil {
			fmt.Printf("   ⚠️  Failed to save image %s: %v\n", outputPath, err)
			continue
		}

		fmt.Printf("   %s (%d points)\n", base, len(points))
	}
}

// loadLabelPoints 加载标签点（完善错误处理+越界校验）
func loadLabelPoints(labelPath string, img image.Image) []LabelPoint {
	data, err := os.ReadFile(labelPath)
	if err != nil {
		return []LabelPoint{}
	}

	var points []LabelPoint
	lines := strings.Split(string(data), "\n")

	for _, li := range lines {
		li = strings.TrimSpace(li)
		if li == "" {
			continue
		}

		fs := strings.Fields(li)
		if len(fs) < 5 {
			continue
		}

		cx, err := strconv.ParseFloat(fs[1], 64)
		if err != nil {
			continue
		}
		cy, err := strconv.ParseFloat(fs[2], 64)
		if err != nil {
			continue
		}

		imgW := img.Bounds().Dx()
		imgH := img.Bounds().Dy()
		x := int(cx*float64(imgW) + 0.5)
		y := int(cy*float64(imgH) + 0.5)

		// 坐标越界校验
		if x < 0 || y < 0 || x > imgW || y > imgH {
			continue
		}

		points = append(points, LabelPoint{X: x, Y: y})
	}

	return points
}

// sortLabelPoints 对标签点排序
func sortLabelPoints(points []LabelPoint) []LabelPoint {
	sorted := make([]LabelPoint, len(points))
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Y != sorted[j].Y {
			return sorted[i].Y < sorted[j].Y
		}
		return sorted[i].X < sorted[j].X
	})
	return sorted
}

// drawAnnotationsOnDC 在已有的 dc 上绘制标注
func drawAnnotationsOnDC(dc *gg.Context, points []LabelPoint, shape string,
	shapeColor, textColor color.Color, shapeSize, textSize, boxSize int, drawText bool) {

	dc.SetFontFace(globalFont)
	sz := float64(boxSize) / 2

	for i, p := range points {
		x, y := float64(p.X), float64(p.Y)

		dc.SetColor(shapeColor)
		dc.SetLineWidth(float64(shapeSize))

		if shape == "rect" {
			dc.DrawRectangle(x-sz, y-sz, float64(boxSize), float64(boxSize))
			dc.Stroke()
		} else {
			dc.DrawCircle(x, y, sz)
			dc.Stroke()
		}

		if drawText {
			dc.SetColor(textColor)
			dc.DrawStringAnchored(strconv.Itoa(i+1), x, y-10, 0.5, 0.5)
		}
	}
}

// ==============================================================================================
// 主函数
// ==============================================================================================

func main() {
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "help":
			fmt.Print(Usage)
			return
		case "version":
			fmt.Println(Version)
			return
		}
	}

	var (
		cmd         = flag.String("cmd", "", "anno|label|pick|label/check")
		csvIn       = flag.String("csv", "", "Input CSV path")
		csvOut      = flag.String("out-csv", "", "Output CSV path")
		shape       = flag.String("shape", "circle", "rect / circle")
		boxSize     = flag.Int("box-size", 88, "Box size")
		input       = flag.String("input", "", "Input CSV path")
		output      = flag.String("output", "", "Output directory")
		imgSrc      = flag.String("image-source", "", "Source image directory")
		checkOut    = flag.String("check-output", "", "Check image output dir")
		checkMode   = flag.String("check-mode", "box", "Check mode")
		shapeColor  = flag.String("shape-color", "yellow", "Shape color")
		shapeSize   = flag.Int("shape-size", 10, "Line width")
		textColor   = flag.String("text-color", "red", "Text color")
		textSize    = flag.Int("text-size", 78, "Font size")
		drawText    = flag.Bool("text", false, "Draw index text")
		sortPts     = flag.Bool("sort", false, "Sort points")
		cShort      = flag.String("c", "", "Short command alias")
		dedup       = flag.Bool("dedup", true, "Enable duplicate removal")
		noDedup     = flag.Bool("no-dedup", false, "Disable duplicate removal")
		showHelp    = flag.Bool("help", false, "Show help")
		showVersion = flag.Bool("version", false, "Show version")
	)

	flag.Usage = func() { fmt.Print(Usage) }
	flag.Parse()

	if *showHelp {
		fmt.Print(Usage)
		return
	}
	if *showVersion {
		fmt.Println(Version)
		return
	}

	realCmd := *cmd
	if *cShort != "" {
		realCmd = *cShort
	}

	if realCmd == "" {
		fmt.Println("Error: No command specified. Use --cmd or -c")
		fmt.Print(Usage)
		os.Exit(1)
	}

	var err error
	switch realCmd {
	case CmdAnno:
		if *csvIn == "" || *csvOut == "" {
			fmt.Println("Error: --csv and --out-csv must be set")
			os.Exit(1)
		}
		enableDedup := *dedup && !*noDedup
		err = RunAnnoCommand(*csvIn, *csvOut, enableDedup)

	case CmdLabel:
		if *input == "" || *output == "" {
			fmt.Println("Error: --input and --output must be set")
			os.Exit(1)
		}
		err = RunLabelCommand(*input, *output, *shape, *boxSize)

	case CmdPick:
		if *input == "" || *output == "" || *imgSrc == "" {
			fmt.Println("Error: --input --output --image-source must be set")
			os.Exit(1)
		}
		err = RunPickCommand(*input, *imgSrc, *output)

	case CmdLabelCheck:
		if *imgSrc == "" || *output == "" || *checkOut == "" {
			fmt.Println("Error: --image-source --output --check-output must be set")
			os.Exit(1)
		}
		// 使用结构体封装参数
		cfg := CheckConfig{
			ImageSource: *imgSrc,
			LabelDir:    *output,
			OutputDir:   *checkOut,
			CheckMode:   *checkMode,
			Shape:       *shape,
			ShapeColor:  ParseColor(*shapeColor),
			TextColor:   ParseColor(*textColor),
			ShapeSize:   *shapeSize,
			TextSize:    *textSize,
			BoxSize:     *boxSize,
			DrawText:    *drawText,
			SortPoints:  *sortPts,
		}
		RunCheckCommand(cfg)
		return

	default:
		fmt.Printf("Error: Unknown command '%s'\n", realCmd)
		fmt.Print(Usage)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}