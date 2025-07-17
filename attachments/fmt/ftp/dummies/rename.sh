#!/bin/sh
clear

# date +%m  
# yields number of month with zero padding - 04 for April
curmonth=$( date +%m )
curyear=$( date +%Y )

# src="/c/Users/pbu/Documents/zew_work/git/other/fmtx-ap/"
src="."
wave1="${curmonth}_${curyear}"

echo "  source dir $src"
echo "  wave       $wave1"


# \cp -p  $src2   ./ftp/
for file in ./*dummy*.pdf; do 
    if [ -f "$file" ]; then 
        file1=${file//dummy/${wave1}}
        echo "    copying $file to $file1 "
        \cp $file "./../${file1}"
    fi 
done


touch ./ZEW_FMT_Expectation_Data.xlsx
cp  -f ./ZEW_FMT_Expectation_Data.xlsx  ../../verkauf/ZEW_FMT_Expectation_Data_dummy.xlsx
echo "expection data dummy - touched and copied"
