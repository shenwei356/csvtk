# Download

`csvtk` is implemented in [Golang](https://golang.org/) programming language,
 executable binary files **for most popular operating system** are freely available
  in [release](https://github.com/shenwei356/csvtk/releases) page.

## Current Version

- [csvtk v0.3.3](https://github.com/shenwei356/csvtk/releases/tag/v0.3.3)
    - fix bug of `csvtk grep -t -P`

## Installation

[Download Page](https://github.com/shenwei356/csvtk/releases)

Just [download](https://github.com/shenwei356/csvtk/releases) gzip-compressed
executable file of your operating system, and uncompress it with `gzip -d *.gz` command,
rename it to `csvtk.exe` (Windows) or `csvtk` (other operating systems) for convenience.

You may need to add executable permision by `chmod a+x csvtk`.

You can also add the directory of the executable file to environment variable
`PATH`, so you can run `csvtk` anywhere.

1. For windows, the simplest way is copy it to `C:\WINDOWS\system32`.

2. For Linux, type:

        chmod a+x /PATH/OF/FASTCOV/csvtk
        echo export PATH=\$PATH:/PATH/OF/FASTCOV >> ~/.bashrc

    or simply copy it to `/usr/local/bin`

## Previous Versions

- [csvtk v0.3.2](https://github.com/shenwei356/csvtk/releases/tag/v0.3.2)
    - fix bug of `inter`
- [csvtk v0.3.1](https://github.com/shenwei356/csvtk/releases/tag/v0.3.1)
    - add support of search multiple fields for `grep`
- [csvtk v0.3](https://github.com/shenwei356/csvtk/releases/tag/v0.3)
    - add subcommand `csv2md`
- [csvtk v0.2.9](https://github.com/shenwei356/csvtk/releases/tag/v0.2.9)
    - add more flags to subcommand `pretty`
    - fix bug of `csvtk cut -n`
    - add subcommand `filter`
- [csvtk v0.2.8](https://github.com/shenwei356/csvtk/releases/tag/v0.2.8)
    - add subcommand `pretty` -- convert CSV to readable aligned table
- [csvtk v0.2.7](https://github.com/shenwei356/csvtk/releases/tag/v0.2.7)
    - fix highlight failing in windows
- [csvtk v0.2.6](https://github.com/shenwei356/csvtk/releases/tag/v0.2.6)
    - fix one error message of `grep`
    - highlight matched fields in result of `grep`
- [csvtk v0.2.5](https://github.com/shenwei356/csvtk/releases/tag/v0.2.5)
    - fix bug of `stat` that failed to considerate files with header row
    - add subcommand `stat2` - summary of selected number fields
    - make the output of `stat` prettier
- [csvtk v0.2.4](https://github.com/shenwei356/csvtk/releases/tag/v0.2.4)
    - fix bug of handling comment lines
    - add some notes before using csvtk
- [csvtk v0.2.3](https://github.com/shenwei356/csvtk/releases/tag/v0.2.3)
    - add flag `--colnames` to `cut`
    - flag `-f` (`--fields`) of `join` supports single value now
- [csvtk v0.2.2](https://github.com/shenwei356/csvtk/releases/tag/v0.2.2)
    - add flag `--keep-unmathed` to `join`
- [csvtk v0.2](https://github.com/shenwei356/csvtk/releases/tag/v0.2)
    - finish almost functions
- [csvtk v0.2.1](https://github.com/shenwei356/csvtk/releases/tag/v0.2.1)
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
