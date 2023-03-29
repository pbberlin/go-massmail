#!/bin/sh
clear

src="/c/Users/pbu/Documents/zew_work/git/other/fmtx-ap/"
wave1="202303"
wave2="2023-03"
wave2="2023-03"
wave3="03_2023"

echo "  source dir $src"
echo "  wave       $wave1"



echo "presse pdf"
src2="${src}Pressemitteilung_${wave1}_??.pdf"
\cp -p  $src2   ./pressemitteilungen/${dst2}
for file in ./pressemitteilungen/*; do 
    if [ -f "$file" ]; then 
        echo "    replacing $wave1 in $file"
        mv $file ${file//${wave1}_/}
    fi 
done

\cp -p  $src2   ./ftp/
for file in ./ftp/Pressemitteilung*.pdf; do 
    if [ -f "$file" ]; then 
        echo "    replacing Pressemitteilung with $wave3 in $file"
        file1=${file//Pressemitteilung_/}
        file2=${file1//$wave1/$wave3}
        file3=${file2//_dt.pdf/.pdf}
        mv $file $file3
    fi 
done
for file in ./ftp/*_en.pdf; do 
    if [ -f "$file" ]; then 
        echo "    replacing _en with e_ in $file"
        file1=${file//$wave3/e_${wave3}}
        mv $file ${file1//_en/}
    fi 
done


# echo "  file       $src2"


echo "konjunktur"
src4="${src}konjunktur.xls"
\cp -p  $src4   ./
\cp -p  $src4   ./ftp/


echo "tabellen-1"
src5="${src}tab*.*"
\cp -p  $src5   ./tabellen/
for file in ./tabellen/*; do 
    if [ -f "$file" ]; then 
        echo "    replacing $wave2 in $file"
        mv $file ${file//-${wave2}/}
    fi 
done

echo "tabellen-ftp"
\cp -p  $src5   ./ftp/
for file in ./ftp/tab*.*; do 
    if [ -f "$file" ]; then 
        echo "    replacing $wave2 with $wave3 in $file"
        mv $file ${file//${wave2}/${wave3}}
    fi 
done
for file in ./ftp/tab-engl*.*; do 
    if [ -f "$file" ]; then 
        echo "    replacing tab-engl with e_*_table.pdf in $file"
        file1=${file//tab-engl-/e_}
        file2=${file1//.pdf/_table.pdf}
        mv $file $file2
    fi 
done
for file in ./ftp/tab-*.*; do 
    if [ -f "$file" ]; then 
        echo "    replacing tab- with *_Tabelle.pdf in $file"
        file1=${file//tab-/}
        file2=${file1//.pdf/_Tabelle.pdf}
        mv $file $file2
    fi 
done



echo "verkauf"
src6="${src}Verkauf/*.xlsx"
\cp -p  $src6   ./verkauf/
for file in ./verkauf/*; do 
    if [ -f "$file" ]; then 
        echo "    replacing .xls.xlsx in $file"
        mv $file ${file//.xls.xlsx/.xlsx}
    fi 
done
