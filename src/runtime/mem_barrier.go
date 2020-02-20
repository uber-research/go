package runtime

// a hardware memory barrier that prevents any memory write access from
// being moved and executed on the other side of the barrier
func mb()

// a hardware memory barrier that prevents any memory read access from
// being moved and executed on the other side of the barrier
func rmb()

// A hardware memory barrier that prevents any memory write access from
// being moved and executed on the other side of the barrier
func wmb()
