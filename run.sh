go build

echo "GET http://localhost:8080/hello" > target.txt

./slapperx -minY 900ms -maxY=10s -rate 200 -workers=2000 -timeout 10s -targets target.txt

#python3 -m termtosvg render ui.asciicast demo.svg