#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

total_packages=0
passed_packages=0
failed_packages=0

echo -e "${YELLOW}Starting Go tests...${NC}\n"

# Test all packages (excluding vendor)
for pkg in $(go list ./... | grep -v /vendor/); do
    ((total_packages++))
    echo -e "${YELLOW}Testing $pkg${NC}"

    go test -cover -v "$pkg"

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}PASS: $pkg${NC}\n"
        ((passed_packages++))
    else
        echo -e "${RED}FAIL: $pkg${NC}\n"
        ((failed_packages++))
    fi
done

# Summary
echo -e "${YELLOW}Total: $total_packages | ${GREEN}Passed: $passed_packages | ${RED}Failed: $failed_packages${NC}"

# Exit code
if [ $failed_packages -gt 0 ]; then
    exit 1
else
    exit 0
fi
