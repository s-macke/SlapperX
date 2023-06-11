set -e

go build

echo "GET http://localhost:8080/hello" > target.txt

#echo "GET http://localhost:8080/product/100001" > target.txt
#echo "###" >> target.txt
#echo "GET http://localhost:8080/product/100002" >> target.txt
#echo "###" >> target.txt
#echo "GET http://localhost:8080/product/100003" >> target.txt
#echo "###" >> target.txt
#echo "GET http://localhost:8080/product/100004" >> target.txt

./slapperx -minY 990ms -maxY=12s -rate 250 -workers=2500 -timeout 10s -targets target.txt
#./slapperx -minY 0ms -maxY=10ms -rate 1000 -workers=2500 -timeout 10s -targets target.txt

#python3 -m termtosvg render ui.asciicast demo.svg
