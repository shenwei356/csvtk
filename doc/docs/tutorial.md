# Tutorial

## Forewords

Yes, you could just use spreadsheet softwares like MS excel to
do most of the job.

Howerver it's all by clicking and typing, which is **not
automatically and time-consuming to repeate**, especially when we want to
apply similar operations with different datasets or purposes.

`csvtk` is **convenient for rapid investigation
and also easy to integrated into analysis pipelines**.
 It could save you much time of writting scripts.

Hope it be helpful for you.

## Analyze OTU table

### Data

Here is mock a OTU table from 16S sequencing result.
Columns are sample IDs in format of "GROUP.ID"

        $ cat otu_table.csv
        Taxonomy,A.1,A.2,A.3,B.1,B.2,B.3,C.1,C.2,C.3
        Proteobacteria,.13,.29,.13,.16,.13,.22,.30,.23,.21
        Firmicutes,.42,.06,.49,.41,.55,.41,.32,.38,.66
        Bacteroidetes,.19,.62,.12,.33,.16,.29,.34,.35,.09
        Deferribacteres,.17,.00,.24,.01,.01,.01,.01,.01,.02
        Tenericutes,.00,.00,.00,.01,.03,.02,.00,.00,.00

### Steps

1. Counting

        $ csvtk stat otu_table.csv
        file: otu_table.csv  num_cols: 10  num_rows: 6

1. Convert to tab-delimited table

        $ csvtk csv2tab  otu_table.csv
        Taxonomy        A.1     A.2     A.3     B.1     B.2     B.3     C.1     C.2     C.3
        Proteobacteria  .13     .29     .13     .16     .13     .22     .30     .23     .21
        Firmicutes      .42     .06     .49     .41     .55     .41     .32     .38     .66
        Bacteroidetes   .19     .62     .12     .33     .16     .29     .34     .35     .09
        Deferribacteres .17     .00     .24     .01     .01     .01     .01     .01     .02
        Tenericutes     .00     .00     .00     .01     .03     .02     .00     .00     .00

1. Extract data of group A and B and save to file `-o otu_table.gAB.csv`

        $ csvtk cut -F -f "A.*,B.*,Taxonomy" otu_table.csv -o otu_table.gAB.csv
        $ cat otu_table.gAB.csv
        Taxonomy,A.1,A.2,A.3,B.1,B.2,B.3
        Proteobacteria,.13,.29,.13,.16,.13,.22
        Firmicutes,.42,.06,.49,.41,.55,.41
        Bacteroidetes,.19,.62,.12,.33,.16,.29
        Deferribacteres,.17,.00,.24,.01,.01,.01
        Tenericutes,.00,.00,.00,.01,.03,.02

1. Transpose

        $ csvtk transpose otu_table.gAB.csv -o otu_table.gAB.t.csv
        $ csvtk csv2tab  otu_table.gAB.t.csv         
        Taxonomy        Proteobacteria  Firmicutes      Bacteroidetes   Deferribacteres Tenericutes
        A.1     .13     .42     .19     .17     .00
        A.2     .29     .06     .62     .00     .00
        A.3     .13     .49     .12     .24     .00
        B.1     .16     .41     .33     .01     .01
        B.2     .13     .55     .16     .01     .03
        B.3     .22     .41     .29     .01     .02

1. Rename first column

        $ csvtk rename -f 1 -n "sample" otu_table.gAB.t.csv -o otu_table.gAB.t.r.csv
        $ csvtk csv2tab  otu_table.gAB.t.r.csv
        sample  Proteobacteria  Firmicutes      Bacteroidetes   Deferribacteres Tenericutes
        A.1     .13     .42     .19     .17     .00
        A.2     .29     .06     .62     .00     .00
        A.3     .13     .49     .12     .24     .00
        B.1     .16     .41     .33     .01     .01
        B.2     .13     .55     .16     .01     .03
        B.3     .22     .41     .29     .01     .02

1. Add group column

        $ csvtk mutate -p "(.+?)\." -n group otu_table.gAB.t.r.csv -o otu_table2.csv
        $ csvtk csv2tab otu_table2.csv
        sample  Proteobacteria  Firmicutes      Bacteroidetes   Deferribacteres Tenericutes     group
        A.1     .13     .42     .19     .17     .00     A
        A.2     .29     .06     .62     .00     .00     A
        A.3     .13     .49     .12     .24     .00     A
        B.1     .16     .41     .33     .01     .01     B
        B.2     .13     .55     .16     .01     .03     B
        B.3     .22     .41     .29     .01     .02     B

1. Rename groups:

        $ csvtk replace -f group -p "A" -r "Ctrl" otu_table2.csv | csvtk replace -f group -p "B" -r "Treatment" > otu_table3.csv
        $ csvtk csv2tab otu_table3.csv sample  Proteobacteria  Firmicutes      Bacteroidetes   Deferribacteres Tenericutes     group
        A.1     .13     .42     .19     .17     .00     Ctrl
        A.2     .29     .06     .62     .00     .00     Ctrl
        A.3     .13     .49     .12     .24     .00     Ctrl
        B.1     .16     .41     .33     .01     .01     Treatment
        B.2     .13     .55     .16     .01     .03     Treatment
        B.3     .22     .41     .29     .01     .02     Treatment


1. Sort by abundance of *Proteobacteria* in descending order.

        $ csvtk sort -k Proteobacteria:nr otu_table3.csv -T
        sample  Proteobacteria  Firmicutes      Bacteroidetes   Deferribacteres Tenericutes     group
        A.2     .29     .06     .62     .00     .00     Ctrl
        B.3     .22     .41     .29     .01     .02     Treatment
        B.1     .16     .41     .33     .01     .01     Treatment
        B.2     .13     .55     .16     .01     .03     Treatment
        A.3     .13     .49     .12     .24     .00     Ctrl
        A.1     .13     .42     .19     .17     .00     Ctrl

1. Sort by abundance of *Proteobacteria* in descending order and *Firmicutes* in ascending order

        $ csvtk sort -k Proteobacteria:nr -k Firmicutes:n otu_table3.csv -T
        sample  Proteobacteria  Firmicutes      Bacteroidetes   Deferribacteres Tenericutes     group
        A.2     .29     .06     .62     .00     .00     Ctrl
        B.3     .22     .41     .29     .01     .02     Treatment
        B.1     .16     .41     .33     .01     .01     Treatment
        A.1     .13     .42     .19     .17     .00     Ctrl
        A.3     .13     .49     .12     .24     .00     Ctrl
        B.2     .13     .55     .16     .01     .03     Treatment


        
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
