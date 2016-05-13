CYAN='\033[0;36m'
NC='\033[0m'
echo "\n"
echo "${CYAN}------------------------------------------------${NC}"
echo "\n"
go build && ./evingo parse examples/clock.e
