set term pngcairo size 1024, 256


set xlabel "time (s)"
set ylabel "log duration (undefined)"
set autoscale xfix
set autoscale yfix
set autoscale cbfix
#set cbrange [0:300]

set output "cs_cz_int.png"
plot [:1500] 'cs_cz_int.dat' matrix nonuniform with image notitle

set output "uk_ua_int.png"
plot [:1500] 'uk_ua_int.dat' matrix nonuniform with image notitle


set output "cs_cz_int_2.png"
plot [:1500] 'cs_cz_int_2.dat' matrix nonuniform with image notitle



set output "cs_cz_qa.png"
plot [:1500] 'cs_cz_qa.dat' matrix nonuniform with image notitle

set output "nl_nl_int.png"
plot [:1500] 'nl_NL_int.dat' matrix nonuniform with image notitle

set output "out.png"
plot [:1500] 'out.dat' matrix nonuniform with image notitle
