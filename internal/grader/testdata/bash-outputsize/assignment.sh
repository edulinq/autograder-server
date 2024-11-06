function add() {
    # Write 1KB to stderr (this gets called 5 times).
    base64 /dev/urandom | head -c 3k 1>&2

    local a=$1
    local b=$2
    echo $((a + b))
}
