Here is the English version of the README file based on your Chinese version:

```markdown
# iufx - Skewer/Bamboo Dataset Annotation Tool
**All-in-one Dataset Annotation Processing Tool** with CSV sorting/deduplication, YOLO format conversion, image filtering, and annotation visualization. Supports local execution + containerized deployment.

## ✨ Core Features
1. **`anno` Smart CSV Sorting & Deduplication** (by filename + Y/X coordinates)
2. **`label` Convert to YOLO Format Labels** (supports rectangle/circle)
3. **`pick` Batch Image Filtering** (auto-copy based on CSV)
4. **`label/check` Annotation Visualization** (draw boxes/circles + indices)
5. **`version` Display version information**
6. **`help` Display full help**


## 📌 Complete Parameter Reference

### Global Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--cmd` | Command: anno/label/pick/label/check | - |
| `--help` | Display help message | - |
| `--version` | Display version number | - |

### anno Command Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--csv` | Input CSV file | - |
| `--out-csv` | Output CSV file | - |
| `--dedup` | Enable duplicate point removal | true |
| `--no-dedup` | Disable duplicate point removal | false |

### label Command Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--input` | Input CSV file | - |
| `--output` | Label output directory | - |
| `--shape` | Shape type (rect/circle) | circle |
| `--box-size` | Annotation box size | 88 |

### pick Command Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--input` | Input CSV file | - |
| `--output` | Image output directory | - |
| `--image-source` | Source image folder | - |

### label/check Command Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--image-source` | Image folder | - |
| `--output` | Label folder | - |
| `--check-output` | Check image output directory | - |
| `--shape` | Shape type (rect/circle) | circle |
| `--box-size` | Annotation box size | 88 |
| `--shape-color` | Shape color | yellow |
| `--shape-size` | Line width | 10 |
| `--text-color` | Text color | red |
| `--text-size` | Font size | 78 |
| `--text` | Enable index text | false |
| `--sort` | Sort points before drawing | false |

### Supported Colors

| Color Name | Description |
|------------|-------------|
| `red` | Red |
| `yellow` | Yellow |
| `blue` | Blue |
| `green` | Green |
| `white` | White |
| `rgba(r,g,b,a)` | Custom RGBA color |


## 🚀 Most Common Command Examples

### 1. CSV Sorting & Deduplication
```bash
./bin/iufx --cmd anno --csv input.csv --out-csv output.csv
./bin/iufx --cmd anno --csv input.csv --out-csv output.csv --no-dedup
```

### 2. Convert to YOLO Labels
```bash
./bin/iufx --cmd label --input annotations.csv --output labels/
./bin/iufx --cmd label --input annotations.csv --output labels/ --shape rect --box-size 100
```

### 3. Batch Image Filtering
```bash
./bin/iufx --cmd pick --input annotations.csv --image-source /data/images --output selected/
```

### 4. Visualization Check
```bash
./bin/iufx --cmd label/check --image-source images/ --output labels/ --check-output checked/ --text
./bin/iufx --cmd label/check --image-source images/ --output labels/ --check-output checked/ --shape rect --shape-color blue --sort
```

### 5. View Version & Help
```bash
./bin/iufx version
./bin/iufx help
```

## 🐳 Containerized Execution (Complete Tutorial)
```sh
# Build executable and export to bin/ directory
# docker build --progress=plain -f Dockerfile.iufx --target export -o bin .
# iufx help

# Build the image
docker build --progress=plain -f Dockerfile.iufx --target runtime -t ymc/iufx .

# Start container with bash
docker run --rm -it -v $(pwd):/app -v /mnt/e/download/babo:/babo ymc/iufx bash
# 

# iufx help
# iufx version
# iufx --help
# iufx --version

# ls /babo

# Sort annotations
iufx --cmd anno --csv /babo/bamboo_annotations.csv --out-csv /babo/bamboo_annotations_sorted.csv
# ls /babo/*.csv

# Generate labels
iufx -c label --shape rect --box-size 88 --input /babo/bamboo_annotations_sorted.csv --output /babo/dataset/labels/train
# ls /babo/dataset/labels/train

# Pick images
iufx -c pick --input /babo/bamboo_annotations_sorted.csv --output /babo/dataset/images/train --image-source /babo/dataset_x
# ls /babo/dataset/images/train

# Check labels
iufx -c label/check --image-source /babo/dataset/images/train --check-output /babo/dataset/images/check --output /babo/dataset/labels/train --check-mode box --shape circle --box-size 88 --shape-color yellow --shape-size 10 --text-color red --text --sort --text-size 78
# ls /babo/dataset/images/check


# rm -r /babo/dataset/images/check

# Fix labels (--box-size xx)
# ...
```

### 4. Containerization Advantages
- No Go environment installation required
- Consistent cross-platform execution
- Secure isolation, no system pollution
- Suitable for CI/CD automation pipelines


## 📄 CSV Format Specification

The input CSV file must contain the following columns (at least 6 columns):

| Index | Field Name | Description | Example |
|-------|------------|-------------|---------|
| 0 | Class | Category name | bamboo |
| 1 | X | X coordinate | 320 |
| 2 | Y | Y coordinate | 240 |
| 3 | Filename | Image filename | image_001.jpg |
| 4 | Width | Image width | 1920 |
| 5 | Height | Image height | 1080 |

### CSV Example
```csv
bamboo,320,240,image_001.jpg,1920,1080
bamboo,450,380,image_001.jpg,1920,1080
bamboo,120,500,image_002.jpg,1920,1080
```


## 📂 Output Description

### anno Output
- Sorted CSV file, ordered by filename → Y coordinate → X coordinate
- Optional removal of duplicate coordinate points

### label Output
- One `.txt` file per image
- YOLO format: `class_id center_x center_y width height`
```
0 0.166667 0.222222 0.045833 0.081481
0 0.234375 0.351852 0.045833 0.081481
```

### pick Output
- Copies image files referenced in CSV (only `.jpg/.jpeg/.png`)

### label/check Output
- PNG images with drawn annotations
- Optional index display, rectangle/circle drawing


## 🛡️ Feature Highlights
- **Coordinate Validation**: Auto-filters negative coordinates and invalid dimensions
- **Auto-deduplication**: anno command supports removing duplicate coordinate points
- **Image Format Validation**: pick command auto-filters non-image files
- **Memory Optimized**: Streaming CSV reading, suitable for large files
- **Boundary Protection**: Auto-validates coordinate boundaries during visualization


## 📊 Workflow Example

```bash
# 1. Sort and deduplicate original CSV
./bin/iufx --cmd anno --csv raw.csv --out-csv sorted.csv

# 2. Generate YOLO labels
./bin/iufx --cmd label --input sorted.csv --output labels/

# 3. Filter relevant images
./bin/iufx --cmd pick --input sorted.csv --image-source /original_images --output selected_images/

# 4. Visualize and check annotation quality
./bin/iufx --cmd label/check --image-source selected_images/ --output labels/ --check-output checked/ --text --sort
```


## 💡 FAQ

### Q: What if the CSV file is very large?
A: The program uses streaming reads and can handle GB-level CSV files.

### Q: How to customize annotation colors?
A: Use `--shape-color rgba(255,128,0,255)` format for custom colors.

### Q: Too many annotation points, indices are hard to see?
A: Adjust the `--text-size` parameter to increase font size, or use `--shape-size` to adjust line thickness.

### Q: Which image formats are supported?
A: Supports `.jpg`, `.jpeg`, and `.png` formats.