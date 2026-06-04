# iufx - 烧烤竹签类数据集标注工具
**一站式数据集标注处理工具**，支持 CSV 排序去重、YOLO 格式转换、图像筛选、标注可视化，支持本地运行 + 容器化部署。


## ✨ 核心功能
1. **`anno` CSV 智能排序去重**（按文件名 + Y/X 坐标）
2. **`label` 转换为 YOLO 格式标签**（支持矩形/圆形）
3. **`pick` 批量筛选图像**（根据 CSV 自动复制）
4. **`label/check` 标注可视化检查**（绘制标注框/圆 + 序号）
5. **`version` 查看版本信息**
6. **`help` 查看完整帮助**


## 📌 完整参数说明

### 全局参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--cmd` | 命令：anno/label/pick/label/check | - |
| `--help` | 显示帮助信息 | - |
| `--version` | 显示版本号 | - |

### anno 命令参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--csv` | 输入 CSV 文件 | - |
| `--out-csv` | 输出 CSV 文件 | - |
| `--dedup` | 启用重复点去重 | true |
| `--no-dedup` | 禁用重复点去重 | false |

### label 命令参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--input` | 输入 CSV 文件 | - |
| `--output` | 标签输出目录 | - |
| `--shape` | 形状类型 (rect/circle) | circle |
| `--box-size` | 标注框大小 | 88 |

### pick 命令参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--input` | 输入 CSV 文件 | - |
| `--output` | 图像输出目录 | - |
| `--image-source` | 源图像文件夹 | - |

### label/check 命令参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--image-source` | 图像文件夹 | - |
| `--output` | 标签文件夹 | - |
| `--check-output` | 检查图像输出目录 | - |
| `--shape` | 形状类型 (rect/circle) | circle |
| `--box-size` | 标注框大小 | 88 |
| `--shape-color` | 形状颜色 | yellow |
| `--shape-size` | 线条宽度 | 10 |
| `--text-color` | 文字颜色 | red |
| `--text-size` | 字体大小 | 78 |
| `--text` | 启用序号文字 | false |
| `--sort` | 绘制前排序点 | false |

### 支持的颜色值

| 颜色名 | 说明 |
|--------|------|
| `red` | 红色 |
| `yellow` | 黄色 |
| `blue` | 蓝色 |
| `green` | 绿色 |
| `white` | 白色 |
| `rgba(r,g,b,a)` | 自定义 RGBA 颜色 |


## 🚀 最常用命令示例

### 1. CSV 排序去重
```bash
./bin/iufx --cmd anno --csv input.csv --out-csv output.csv
./bin/iufx --cmd anno --csv input.csv --out-csv output.csv --no-dedup
```

### 2. 转换为 YOLO 标签
```bash
./bin/iufx --cmd label --input annotations.csv --output labels/
./bin/iufx --cmd label --input annotations.csv --output labels/ --shape rect --box-size 100
```

### 3. 批量筛选图像
```bash
./bin/iufx --cmd pick --input annotations.csv --image-source /data/images --output selected/
```

### 4. 可视化检查标注
```bash
./bin/iufx --cmd label/check --image-source images/ --output labels/ --check-output checked/ --text
./bin/iufx --cmd label/check --image-source images/ --output labels/ --check-output checked/ --shape rect --shape-color blue --sort
```

### 5. 查看版本和帮助
```bash
./bin/iufx version
./bin/iufx help
```

## 🐳 容器化运行（完整教程）
```sh
# 构建可执行文件 并导出到 bin/ 目录下
# docker build --progress=plain -f Dockerfile.iufx --target export -o bin .
# iufx help

# 构建镜像
docker build --progress=plain -f Dockerfile.iufx --target runtime -t ymc/iufx .

# 登录容器 bash
docker run --rm -it -v $(pwd):/app -v /mnt/e/download/babo:/babo ymc/iufx bash
# 

# iufx help
# iufx version
# iufx --help
# iufx --version

# ls /babo

# 排序 annotations
iufx --cmd anno --csv /babo/bamboo_annotations.csv --out-csv /babo/bamboo_annotations_sorted.csv
# ls /babo/*.csv

# 生成 labels
iufx -c label --shape rect --box-size 88 --input /babo/bamboo_annotations_sorted.csv --output /babo/dataset/labels/train
# ls /babo/dataset/labels/train

# 挑选 images
iufx -c pick --input /babo/bamboo_annotations_sorted.csv --output /babo/dataset/images/train --image-source /babo/dataset_x
# ls /babo/dataset/images/train

# 检查 labels
iufx -c label/check --image-source /babo/dataset/images/train --check-output /babo/dataset/images/check --output /babo/dataset/labels/train --check-mode box --shape circle --box-size 88 --shape-color yellow --shape-size 10 --text-color red --text --sort --text-size 78
# ls /babo/dataset/images/check


# rm -r /babo/dataset/images/check

# 修正 labels (--box-size xx)
# ...

```

### 容器化优势
- 无需安装 Go 环境
- 跨平台一致运行
- 安全隔离，不污染系统
- 适合 CI/CD 自动化流程


## 📄 CSV 格式说明

输入 CSV 文件需包含以下列（至少 6 列）：

| 列索引 | 字段名 | 说明 | 示例 |
|--------|--------|------|------|
| 0 | Class | 类别名称 | bamboo |
| 1 | X | X 坐标 | 320 |
| 2 | Y | Y 坐标 | 240 |
| 3 | Filename | 图像文件名 | image_001.jpg |
| 4 | Width | 图像宽度 | 1920 |
| 5 | Height | 图像高度 | 1080 |

### CSV 示例
```csv
bamboo,320,240,image_001.jpg,1920,1080
bamboo,450,380,image_001.jpg,1920,1080
bamboo,120,500,image_002.jpg,1920,1080
```


## 📂 输出说明

### anno 输出
- 排序后的 CSV 文件，按文件名 → Y 坐标 → X 坐标排序
- 可选去除重复坐标点

### label 输出
- 每张图像对应一个 `.txt` 文件
- YOLO 格式：`class_id center_x center_y width height`
```
0 0.166667 0.222222 0.045833 0.081481
0 0.234375 0.351852 0.045833 0.081481
```

### pick 输出
- 复制 CSV 中引用的图像文件（仅 `.jpg/.jpeg/.png`）

### label/check 输出
- 绘制标注后的 PNG 图像
- 可选显示序号、矩形/圆形标注


## 🛡️ 特性说明
- **坐标合法性校验**：自动过滤负数坐标和无效尺寸
- **自动去重**：anno 命令支持去除相同坐标的重复点
- **图片格式校验**：pick 命令自动过滤非图片文件
- **内存优化**：流式读取 CSV，适合大文件处理
- **越界保护**：可视化时自动校验坐标边界


## 📊 工作流程示例

```bash
# 1. 原始 CSV 排序去重
./bin/iufx --cmd anno --csv raw.csv --out-csv sorted.csv

# 2. 生成 YOLO 标签
./bin/iufx --cmd label --input sorted.csv --output labels/

# 3. 筛选相关图像
./bin/iufx --cmd pick --input sorted.csv --image-source /original_images --output selected_images/

# 4. 可视化检查标注质量
./bin/iufx --cmd label/check --image-source selected_images/ --output labels/ --check-output checked/ --text --sort
```


## 💡 常见问题

### Q: CSV 文件很大怎么办？
A: 程序使用流式读取，可以处理 GB 级别的 CSV 文件。

### Q: 如何自定义标注颜色？
A: 使用 `--shape-color rgba(255,128,0,255)` 格式自定义颜色。

### Q: 标注点太多看不清序号？
A: 可以调整 `--text-size` 参数增大字体，或使用 `--shape-size` 调整线条粗细。

### Q: 支持哪些图像格式？
A: 支持 `.jpg`、`.jpeg`、`.png` 格式。

