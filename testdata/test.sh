#!/bin/bash

test -e ssshtest || wget -q https://raw.githubusercontent.com/ryanlayer/ssshtest/master/ssshtest

. ssshtest
set -e


cd csvtk; go build; cd ..;
app=./csvtk/csvtk

set +e


# ----------------------------------------------------------------------------
# csvtk headers
# ----------------------------------------------------------------------------

for n in 1 10 100 10000 1000000; do
    fn() {
        cat <(seq $n | awk '{print "c"$1}' | paste -sd,) <(seq $n | paste -sd,) \
            | $app headers
    }
    run "headers (n=$n)" fn
    assert_no_stderr
    assert_equal $(cat $STDOUT_FILE | grep -v '#' | wc -l) $n
done

# ----------------------------------------------------------------------------
# csvtk stats
# ----------------------------------------------------------------------------

for n in 1 10 100 10000 1000000; do
    fn() {
        cat <(seq $n | awk '{print "c"$1}' | paste -sd,) <(seq $n | paste -sd,) \
            | $app stats
    }
    run "stats (n=$n)" fn
    assert_no_stderr
    assert_equal $(cat $STDOUT_FILE | $app space2tab | sed 1d | cut -f 2 | sed 's/,//g') $n
    assert_equal $(cat $STDOUT_FILE | $app space2tab | sed 1d | cut -f 3 | sed 's/,//g') 1

    fn() {
        cat <(seq $n | awk '{print "c"$1}' | paste -sd,) <(seq $n | paste -sd,) \
            | $app stats -H
    }
    run "stats -H (n=$n)" fn
    assert_no_stderr
    assert_equal $(cat $STDOUT_FILE | $app space2tab | sed 1d | cut -f 2 | sed 's/,//g') $n
    assert_equal $(cat $STDOUT_FILE | $app space2tab | sed 1d | cut -f 3 | sed 's/,//g') 2

    # -----------------------------------------------------------------------
    # transpose
    # -----------------------------------------------------------------------
    fn() {
        cat <(seq $n | awk '{print "c"$1}' | paste -sd,) <(seq $n | paste -sd,) \
            | $app transpose \
            | $app stats
    }
    run "transpose & stats (n=$n)" fn
    assert_no_stderr
    assert_equal $(cat $STDOUT_FILE | $app space2tab | sed 1d | cut -f 2 | sed 's/,//g') 2
    assert_equal $(cat $STDOUT_FILE | $app space2tab | sed 1d | cut -f 3 | sed 's/,//g') $(($n -1))

    fn() {
        cat <(seq $n | awk '{print "c"$1}' | paste -sd,) <(seq $n | paste -sd,) \
            | $app transpose \
            | $app stats -H
    }
    run "transpose & stats -H (n=$n)" fn
    assert_no_stderr
    assert_equal $(cat $STDOUT_FILE | $app space2tab | sed 1d | cut -f 2 | sed 's/,//g') 2
    assert_equal $(cat $STDOUT_FILE | $app space2tab | sed 1d | cut -f 3 | sed 's/,//g') $n
done


# ----------------------------------------------------------------------------
# csvtk cut
# ----------------------------------------------------------------------------

for n in 1 10 100 10000 1000000; do
    fn() {
        seq $n | $app cut -H -f 1
    }
    run "cut -H -f 1 (n=$n)"  fn
    assert_no_stderr
    assert_equal $(cat $STDOUT_FILE | wc -l) $n

    fn() {
        seq $n | $app cut -H -t -f 1
    }
    run "cut -H -t -f 1 (n=$n)"  fn
    assert_no_stderr
    assert_equal $(cat $STDOUT_FILE | wc -l) $n

    fn() {
        cat <(echo head) <(seq $n) | $app cut -f head
    }
    run "cut -f head (n=$n)"  fn
    assert_no_stderr
    assert_equal $(cat $STDOUT_FILE | wc -l) $(($n+1))
done
