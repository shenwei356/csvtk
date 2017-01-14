# Download

`csvtk` is implemented in [Go](https://golang.org/) programming language,
 executable binary files **for most popular operating system** are freely available
  in [release](https://github.com/shenwei356/csvtk/releases) page.

## Current Version

[csvtk v0.4.4](https://github.com/shenwei356/csvtk/releases/tag/v0.4.4)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.4.4/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.4)

- add command `csvtk freq`: frequencies of selected fields
- add lots of examples in [usage page](http://bioinf.shenwei.me/csvtk/usage/)


Links:

OS     |Arch      |File                                                                                                                             |Download Count
:------|:---------|:--------------------------------------------------------------------------------------------------------------------------------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
Linux  |32-bit    |[csvtk_linux_386.tar.gz](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_linux_386.tar.gz)                    |[![Github Releases (by Asset)](https://img.shields.io/github/downloads/shenwei356/csvtk/latest/csvtk_linux_386.tar.gz.svg?maxAge=3600)](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_linux_386.tar.gz)
Linux  |**64-bit**|[**csvtk_linux_amd64.tar.gz**](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_linux_amd64.tar.gz)            |[![Github Releases (by Asset)](https://img.shields.io/github/downloads/shenwei356/csvtk/latest/csvtk_linux_amd64.tar.gz.svg?maxAge=3600)](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_linux_amd64.tar.gz)
Linux  |ARM       |[csvtk_linux_arm.tar.gz](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_linux_arm.tar.gz)                    |[![Github Releases (by Asset)](https://img.shields.io/github/downloads/shenwei356/csvtk/latest/csvtk_linux_arm.tar.gz.svg?maxAge=3600)](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_linux_arm.tar.gz)
Linux  |ARM64     |[csvtk_linux_arm64.tar.gz](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_linux_arm64.tar.gz)                |[![Github Releases (by Asset)](https://img.shields.io/github/downloads/shenwei356/csvtk/latest/csvtk_linux_arm64.tar.gz.svg?maxAge=3600)](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_linux_arm64.tar.gz)
OS X   |32-bit    |[csvtk_darwin_386.tar.gz](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_darwin_386.tar.gz)                  |[![Github Releases (by Asset)](https://img.shields.io/github/downloads/shenwei356/csvtk/latest/csvtk_darwin_386.tar.gz.svg?maxAge=3600)](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_darwin_386.tar.gz)
OS X   |**64-bit**|[**csvtk_darwin_amd64.tar.gz**](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_darwin_amd64.tar.gz)          |[![Github Releases (by Asset)](https://img.shields.io/github/downloads/shenwei356/csvtk/latest/csvtk_darwin_amd64.tar.gz.svg?maxAge=3600)](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_darwin_amd64.tar.gz)
Windows|32-bit    |[csvtk_windows_386.exe.tar.gz](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_windows_386.exe.tar.gz)        |[![Github Releases (by Asset)](https://img.shields.io/github/downloads/shenwei356/csvtk/latest/csvtk_windows_386.exe.tar.gz.svg?maxAge=3600)](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_windows_386.exe.tar.gz)
Windows|**64-bit**|[**csvtk_windows_amd64.exe.tar.gz**](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_windows_amd64.exe.tar.gz)|[![Github Releases (by Asset)](https://img.shields.io/github/downloads/shenwei356/csvtk/latest/csvtk_windows_amd64.exe.tar.gz.svg?maxAge=3600)](https://github.com/shenwei356/csvtk/releases/download/v0.4.4/csvtk_windows_amd64.exe.tar.gz)

## Installation

[Download Page](https://github.com/shenwei356/csvtk/releases)

`csvtk` is implemented in [Go](https://golang.org/) programming language,
 executable binary files **for most popular operating systems** are freely available
  in [release](https://github.com/shenwei356/csvtk/releases) page.

Just [download](https://github.com/shenwei356/csvtk/releases) compressed
executable file of your operating system,
and uncompress it with `tar -zxvf xxx.tar.gz` command or other tools.
And then:

1. For Unix-like systems
    1. If you have root privilege simply copy it to `/usr/local/bin`:

            sudo cp csvtk /usr/local/bin/

    1. Or add the current directory of the executable file to environment variable
    `PATH`:

            echo export PATH=\"$(pwd)\":\$PATH >> ~/.bashrc
            source ~/.bashrc

1. **For windows**, just copy `csvtk.exe` to `C:\WINDOWS\system32`.

For Go developer, just one command:

    go get -u github.com/shenwei356/csvtk/csvtk

## Previous Versions

- [csvtk v0.4.3](https://github.com/shenwei356/csvtk/releases/tag/v0.4.3)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.4.3/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.3)
    - improvement of using experience: flag `-n` is not required anymore when flag `-H` in `csvtk mutate`
- [csvtk v0.4.2](https://github.com/shenwei356/csvtk/releases/tag/v0.4.2)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.4.2/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.2)
    - fix highlight bug of `csvtk grep`: if the pattern matches multiple parts,
    the text will be wrongly edited.
    - changes: disable highlight when pattern file given.
    - change the default output of all ploting commands to STDOUT, now you can
    pipe the image to "display" command of Imagemagic.
- [csvtk v0.4.1](https://github.com/shenwei356/csvtk/releases/tag/v0.4.1)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.4.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.1)
    - Nothing changed. Just fix the links due to inappropriate deployment of v0.4.0
- [csvtk v0.4.0](https://github.com/shenwei356/csvtk/releases/tag/v0.4.0)
[![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.4.0/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.0)
    - add flag for `csvtk replace`: `-K` (`--keep-key`) keep the key as value when
    no value found for the key. This is open in default in previous versions.
- [csvtk v0.3.9](https://github.com/shenwei356/csvtk/releases/tag/v0.3.9)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.9/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.4.0)
    - fix bug: header row incomplete in `csvtk sort` result
- [csvtk v0.3.8.1](https://github.com/shenwei356/csvtk/releases/tag/v0.3.8.1)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.8.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.8.1)
    - fix bug of flag parsing library [pflag](https://github.com/spf13/pflag),
    [detail](https://github.com/spf13/pflag/pull/98).
    The bug affected the `csvtk grep -r -p`, when value of `-p` contain "[" and "]"
    at the beginning or end, they are wrongly parsed.
- [csvtk v0.3.8](https://github.com/shenwei356/csvtk/releases/tag/v0.3.8)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.8/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.8)
    - new feature: `csvtk cut` supports ordered fields output. e.g., `csvtk cut -f 2,1`
      outputs the 2nd column in front of 1th column.
    - new commands: `csvtk plot` can plot three types of plots by subcommands:
        - `csvtk plot hist`: histogram
        - `csvtk plot box`: boxplot
        - `csvtk plot line`: line plot and scatter plot
- [csvtk v0.3.7](https://github.com/shenwei356/csvtk/releases/tag/v0.3.7)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.7/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.7)
    - fix a serious bug of using negative field of column name, e.g. `-f "-id"`
- [csvtk v0.3.6](https://github.com/shenwei356/csvtk/releases/tag/v0.3.6)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.6/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.6)
    - `csvtk replace` support replacement symbols `{nr}` (record number)
      and `{kv}` (corresponding value of the key ($1) by key-value file)
- [csvtk v0.3.5.2](https://github.com/shenwei356/csvtk/releases/tag/v0.5.2)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.5.2/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.5.2)
    - add flag `--fill` for `csvtk join`, so we can fill the unmatched data
    - fix typo
- [csvtk v0.3.5.1](https://github.com/shenwei356/csvtk/releases/tag/v0.3.5.1)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.5.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.5.1)
    - fix minor bug of reading lines ending with `\r\n` from a dependency package
- [csvtk v0.3.5](https://github.com/shenwei356/csvtk/releases/tag/v0.3.5)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.5/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.5)
    - fix minor bug of `csv2md`
    - add subcommand `version` which could check for update
- [csvtk v0.3.4](https://github.com/shenwei356/csvtk/releases/tag/v0.3.4)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.4/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.4)
    - fix bug of `csvtk replace` that head row should not be edited.
- [csvtk v0.3.3](https://github.com/shenwei356/csvtk/releases/tag/v0.3.3)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.3/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.3)
    - fix bug of `csvtk grep -t -P`
- [csvtk v0.3.2](https://github.com/shenwei356/csvtk/releases/tag/v0.3.2)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.2/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.2)
    - fix bug of `inter`
- [csvtk v0.3.1](https://github.com/shenwei356/csvtk/releases/tag/v0.3.1)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3.1)
    - add support of search multiple fields for `grep`
- [csvtk v0.3](https://github.com/shenwei356/csvtk/releases/tag/v0.3)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.3/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.3)
    - add subcommand `csv2md`
- [csvtk v0.2.9](https://github.com/shenwei356/csvtk/releases/tag/v0.2.9)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.9/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.9)
    - add more flags to subcommand `pretty`
    - fix bug of `csvtk cut -n`
    - add subcommand `filter`
- [csvtk v0.2.8](https://github.com/shenwei356/csvtk/releases/tag/v0.2.8)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.8/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.8)
    - add subcommand `pretty` -- convert CSV to readable aligned table
- [csvtk v0.2.7](https://github.com/shenwei356/csvtk/releases/tag/v0.2.7)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.7/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.7)
    - fix highlight failing in windows
- [csvtk v0.2.6](https://github.com/shenwei356/csvtk/releases/tag/v0.2.6)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.6/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.6)
    - fix one error message of `grep`
    - highlight matched fields in result of `grep`
- [csvtk v0.2.5](https://github.com/shenwei356/csvtk/releases/tag/v0.2.5)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.5/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.5)
    - fix bug of `stat` that failed to considerate files with header row
    - add subcommand `stat2` - summary of selected number fields
    - make the output of `stat` prettier
- [csvtk v0.2.4](https://github.com/shenwei356/csvtk/releases/tag/v0.2.4)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.4/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.4)
    - fix bug of handling comment lines
    - add some notes before using csvtk
- [csvtk v0.2.3](https://github.com/shenwei356/csvtk/releases/tag/v0.2.3)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.3/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.3)
    - add flag `--colnames` to `cut`
    - flag `-f` (`--fields`) of `join` supports single value now
- [csvtk v0.2.2](https://github.com/shenwei356/csvtk/releases/tag/v0.2.2)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.2/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.2)
    - add flag `--keep-unmathed` to `join`
- [csvtk v0.2](https://github.com/shenwei356/csvtk/releases/tag/v0.2)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2)
    - finish almost functions
- [csvtk v0.2.1](https://github.com/shenwei356/csvtk/releases/tag/v0.2.1)
  [![Github Releases (by Release)](https://img.shields.io/github/downloads/shenwei356/csvtk/v0.2.1/total.svg)](https://github.com/shenwei356/csvtk/releases/tag/v0.2.1)
    - fix bug of `mutate`


<div id="disqus_thread"></div>
<script>
/**
* RECOMMENDED CONFIGURATION VARIABLES: EDIT AND UNCOMMENT THE SECTION BELOW TO INSERT DYNAMIC VALUES FROM YOUR PLATFORM OR CMS.
* LEARN WHY DEFINING THESE VARIABLES IS IMPORTANT: https://disqus.com/admin/universalcode/#configuration-variables
*/
/*
var disqus_config = function () {
this.page.url = PAGE_URL; // Replace PAGE_URL with your page's canonical URL variable
this.page.identifier = PAGE_IDENTIFIER; // Replace PAGE_IDENTIFIER with your page's unique identifier variable
};
*/
(function() { // DON'T EDIT BELOW THIS LINE
var d = document, s = d.createElement('script');

s.src = '//csvtk.disqus.com/embed.js';

s.setAttribute('data-timestamp', +new Date());
(d.head || d.body).appendChild(s);
})();
</script>
<noscript>Please enable JavaScript to view the <a href="https://disqus.com/?ref_noscript" rel="nofollow">comments powered by Disqus.</a></noscript>
