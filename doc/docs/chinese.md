如同生物信息领域中的FASTA/Q格式一样，CSV/TSV作为计算机、数据科学和生物信息的基本格式，应用非常广泛。常用的处理软件包括：

1. 以微软Excel为代表的电子表格软件
2. Notepad++/SublimeText等文本编辑器
3. sed/awk/cut等Shell命令
4. 各种编程语言的数据处理库。

然而，电子表格软件和文本编辑器固然强大，但依赖鼠标操作，不适合批量处理；**sed/awk/cut等Shell命令主要用于通用的表格数据，不适合含有标题行的CSV格式**；**为了一个小操作写Python/R脚本也有点小题大作，且难以复用**。

开发csvtk前现有的工具主要是Python写的csvkit，Rust写的xsv，C语言写的miller，都各有优劣。当时我刚开发完seqkit，投文章过程中时间充足，便想趁热再造一个轮子。

所以我决定写一个命令行工具来满足CSV/TSV格式的常见操作，这就是csvtk了。

## 介绍

基本信息

- 工具类型: 命令行工具，子命令结构
- 支持格式: CSV/TSV, plain/gzip-compressed
- 编程语言: Go
- 支持平台: Linux, OS X， Windows 等
- 发布方式: 单一可执行二进制文件，下载即用
- 发布平台: Github, Bioconda
- 项目主页: http://bioinf.shenwei.me/csvtk/
- 开源地址: https://github.com/shenwei356/csvtk

特性

- 跨平台
- **轻量，无任何依赖，无需编译、配置，下载即用**
- **快速**
- **支持stdin和gzip压缩的输入和输出文件，便于流处理**
- **27个子命令提供多种实用的功能，且能通过命令行管道组合**
- 支持Bash自动补全
- 支持简单的绘图

## 功能

在开发csvtk之前的两三年间，我已经写了几个可以复用的Python/Perl脚本（https://github.com/shenwei356/datakit） ，包括csv2tab、csvtk_grep、csv_join、csv_melt，intersection，unique。所以我的计划是首先集成这些已有的功能，随后根据需求进行扩展。

到目前为止，csvtk已有27个子命令，分为以下几大类：

- 信息
    -  `headers` 直观打印标题行（**操作列数较多的CSV前使用最佳**）
    -  `stats` 基本统计
    -  `stats2` 对选定的数值列进行基本统计
- 格式转化
    -  `pretty` 转为美观、可读性强的格式（**最常用命令之一**）
    -  `csv2tab` 转CSV为制表符分割格式（TSV）
    -  `tab2csv` 转TSV为CSV
    -  `space2tab` 转空格分割格式为TSV
    -  `transpose` 转置CSV/TSV
    -  `csv2md` 转CSV/TSV为makrdown格式（**写文档常用**）
- 集合操作
    -  `head` 打印前N条记录
    -  `sample` 按比例随机采样
    -  `cut` 选择特定列，支持**按列或列名进行基本选择、范围选择、模糊选择、负向选择**（**最常用命令之一，非常强大**）
    -  `uniq` 无须排序，返回按指定（多）列作为key的唯一记录（好绕。。）
    -  `freq` 按指定（多）列进行计数（**常用**）
    -  `inter` 多个文件间的交集
    -  `grep` 指定（多）列为Key进行搜索（**最常用命令之一，可按指定列搜索**）
    -  `filter` 按指定（多）列的数值进行过滤
    -  `filter2` 用类似awk的数值/表达式，按指定（多）列的数值进行过滤
    -  `join` 合并多个文件（**常用**）
- 编辑
    -  `rename` 直接重命名指定（多）列名（**简单而实用**）
    -  `rename2` 以正则表达式重命名指定（多）列名（**简单而实用**）
    -  `replace` 以正则表达式对指定（多）列进行替换编辑（**最常用命令之一，可按指定列编辑**）
    -  `mutate` 以正则表达式基于已有列创建新的一列（**常用于生成多列测试数据**）
    -  `mutate2` 用类似awk的数值/表达式，以正则表达式基于已有（多）列创建新的一列（**常用**）
    -  `gather` 类似于R里面tidyr包的gather方法
- 排序
    -  `sort` 按指定（多）列进行排序
- 绘图
    - `plot` 基本绘图
        - `plot hist` histogram
        - `plot box` boxplot
        - `plot line` line plot and scatter plot
- 其它
    - `version`   版本信息和检查新版本
    - `genautocomplete` 生成支持Bash自动补全的配置文件，重启Terminal生效。

## 使用

1. **输入数据要求每行的列数一致，空行也会报错**
1. **csvtk默认输入数据含有标题行，如没有请开启全局参数`-H`**
1. **csvtk默认输入数据为CSV格式，如为TSV请开启全局参数`-t`**
1. 输入数据列名最好唯一无重复
1. 如果TSV中存在双引号`""`，请开启全局参数`-l`
1. csvtk默认以`#`开始的为注释行，若标题行含`#`，请给全局参数`-C`指定另一个不常见的字符（如`$`）

## 例子

仅提供少量例子，更多例子请看使用手册 http://bioinf.shenwei.me/csvtk/usage/ 。

1. 示例数据

        $ cat names.csv
        id,first_name,last_name,username
        11,"Rob","Pike",rob
        2,Ken,Thompson,ken
        4,"Robert","Griesemer","gri"
        1,"Robert","Thompson","abc"
        NA,"Robert","Abel","123"

1. 增强可读性

        $ cat names.csv  | csvtk pretty
        id   first_name   last_name   username
        11   Rob          Pike        rob
        2    Ken          Thompson    ken
        4    Robert       Griesemer   gri
        1    Robert       Thompson    abc
        NA   Robert       Abel        123

1. 转为markdown

        $ cat names.csv | csvtk csv2md
        id |first_name|last_name|username
        :--|:---------|:--------|:-------
        11 |Rob       |Pike     |rob
        2  |Ken       |Thompson |ken
        4  |Robert    |Griesemer|gri
        1  |Robert    |Thompson |abc
        NA |Robert    |Abel     |123

    效果

    id |first_name|last_name|username
    :--|:---------|:--------|:-------
    11 |Rob       |Pike     |rob
    2  |Ken       |Thompson |ken
    4  |Robert    |Griesemer|gri
    1  |Robert    |Thompson |abc
    NA |Robert    |Abel     |123

1. 用列或列名来选择指定列，可改变列的顺序

        $ cat names.csv | csvtk cut -f 3,1          | csvtk pretty
        $ cat names.csv | csvtk cut -f last_name,id | csvtk pretty
        last_name   id
        Pike        11
        Thompson    2
        Griesemer   4
        Thompson    1
        Abel        NA

1. 用通配符选择多列

        $ cat names.csv | csvtk cut -F -f '*name,id' | csvtk pretty
        first_name   last_name   username   id
        Rob          Pike        rob        11
        Ken          Thompson    ken        2
        Robert       Griesemer   gri        4
        Robert       Thompson    abc        1
        Robert       Abel        123        NA

1. 删除第2，3列（**下列第二种方法是选定范围，但-3在前,-2在后**）

        $ cat names.csv | csvtk cut -f -2,-3                  | csvtk pretty
        $ cat names.csv | csvtk cut -f -3--2                  | csvtk pretty
        $ cat names.csv | csvtk cut -f -first_name,-last_name | csvtk pretty
        id   username
        11   rob
        2    ken
        4    gri
        1    abc
        NA   123

1. 按指定列搜索，**默认精确匹配**

        $ cat names.csv | csvtk grep -f id -p 1 | csvtk pretty
        id   first_name   last_name   username
        1    Robert       Thompson    abc

1. 模糊搜索（正则表达式）

        $ cat names.csv | csvtk grep -f id -p 1 -r | csvtk pretty
        id   first_name   last_name   username
        11   Rob          Pike        rob
        1    Robert       Thompson    abc

1. 用文件作为模式来源

        $ cat names.csv | csvtk grep -f id -P id-files.txt

1. 对指定列做简单替换

        $ cat names.csv | csvtk replace -f id -p '(\d+)' -r 'ID: $1' \
            | csvtk pretty
        id       first_name   last_name   username
        ID: 11   Rob          Pike        rob
        ID: 2    Ken          Thompson    ken
        ID: 4    Robert       Griesemer   gri
        ID: 1    Robert       Thompson    abc
        NA       Robert       Abel        123

1. 用key-value文件来替换（seqkit和brename都支持类似操作）

        $ cat data.tsv
        name    id
        A       ID001
        B       ID002
        C       ID004

        $ cat alias.tsv
        001     Tom
        002     Bob
        003     Jim

        $ csvtk replace -t -f 2 -p "ID(.+)" -r "N: {nr}, alias: {kv}" -k \
            alias.tsv data.tsv
        name    id
        A       N: 1, alias: Tom
        B       N: 2, alias: Bob
        C       N: 3, alias: 004

1. 合并表格，需要分别指定各文件中的key列：默认均为第一列；若列（名）相同提供一个；若不同用分号分割

        $ cat testdata/phones.csv
        username,phone
        gri,11111
        rob,12345
        ken,22222
        shenwei,999999

        $ csvtk join -f 'username;username' --keep-unmatched names.csv phones.csv \
            | csvtk pretty
        id   first_name   last_name   username   phone
        11   Rob          Pike        rob        12345
        2    Ken          Thompson    ken        22222
        4    Robert       Griesemer   gri        11111
        1    Robert       Thompson    abc
        NA   Robert       Abel        123
