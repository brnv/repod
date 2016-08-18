if [[ -f blankd.log ]]; then
    cat blankd.log
fi

if [[ -f blankd.pid ]]; then
    kill -9 $(cat blankd.pid)
fi
