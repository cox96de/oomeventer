SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
sudo docker run -v $SCRIPT_DIR:/work -m 20m python:3.7 python /work/test.py