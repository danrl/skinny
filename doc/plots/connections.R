# Skinny connections over quorum size

without_leader <- function(x) {
	x * (x - 1)
}

with_leader <- function(x) {
	x - 1
}

x <- seq(1,25,2)

plot(x, without_leader(x),
	type="b", col="blue",
	xlab="Quorum Size", ylab="Connections"
)

lines(x, with_leader(x),
	type="b", col = "red"
)

legend("topleft",
	c("Without Leader", "With Leader"),
	fill=c("blue", "red")
)

# minor tickmarks (optional)
library(Hmisc)
minor.tick(nx=5, ny=10, tick.ratio=0.5)
