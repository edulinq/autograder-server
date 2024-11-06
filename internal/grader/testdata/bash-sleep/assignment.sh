function add() {
    # Return the correct answer, but sleep first.
    sleep 5

    local a=$1
    local b=$2
    echo $((a + b))
}
