package bls12381

func ecMapG1(u, t, s *fe) (*fe, *fe) {
	params := eccParamsForG1

	// Evaluate initial point in conic X(u,t),Y(u,t)
	var X1 = new(fe)
	square(X1, u)
	mul(X1, X1, u)
	var Y1 = new(fe)
	square(Y1, t)
	add(X1, X1, params.b)
	sub(X1, X1, Y1)
	add(Y1, Y1, Y1)
	add(Y1, Y1, X1)
	var Z1 = new(fe)
	mul(Z1, u, params.X[0])
	mul(X1, X1, Z1)
	mul(Z1, Z1, t)
	add(Z1, Z1, Z1)

	// Compute projective point in surface S
	//   y = (2Y1)^2
	//   v = X1*Z1 - u*Y1*Z1
	//   w = 2*Y1*Z1
	var y, v, w = new(fe), new(fe), new(fe)
	add(y, Y1, Y1)
	square(y, y)
	mul(v, Y1, u)
	sub(v, X1, v)
	mul(v, v, Z1)
	mul(w, Y1, Z1)
	add(w, w, w)

	// Compute affine point in V
	//   x1 = v/w
	//   x2 = -u-v/w
	//   x3 = u + y^2/w^2
	var x1, x2, x3 *fe = new(fe), new(fe), new(fe)
	inverse(w, w)
	mul(x1, v, w)
	add(x2, u, x1)
	neg(x2, x2)
	mul(x3, y, w)
	square(x3, x3)
	add(x3, u, x3)

	// Compute g(x_i)
	var y21, y22, y23 *fe = new(fe), new(fe), new(fe)
	square(y21, x1)
	add(y21, y21, params.a)
	mul(y21, y21, x1)
	add(y21, y21, params.b)

	square(y22, x2)
	add(y22, y22, params.a)
	mul(y22, y22, x2)
	add(y22, y22, params.b)

	square(y23, x3)
	add(y23, y23, params.a)
	mul(y23, y23, x3)
	add(y23, y23, params.b)

	// Find the square

	x, y := new(fe), new(fe)
	if !isQuadraticNonResidue(y22) { // c2
		x.set(x2)
		y.set(y22)
	} else {
		x.set(x1)
		y.set(y21)
	}

	if !isQuadraticNonResidue(y23) { // c3
		x.set(x3)
		y.set(y23)
	}

	// Find the square-root and choose sign
	sqrt(y, y)
	neg(y22, y)
	if y.sign() != s.sign() { // c1
		y.set(y22) // TODO non-constant time
	}
	return x, y
}

var eccParamsForG1 = struct {
	X   [3]*fe
	Y   [2]*fe
	ax4 *fe
	a   *fe
	b   *fe
}{
	[3]*fe{
		{2156217304866103074, 13022645835963610199, 2695784996601418374, 10964977287082494396, 9032217190667614323, 444895596260081849},
		{1730508156817200468, 9606178027640717313, 7150789853162776431, 7936136305760253186, 15245073033536294050, 1728177566264616342},
		new(fe)},
	[2]*fe{new(fe), new(fe)},
	new(fe),
	new(fe),
	&fe{12260768510540316659, 6038201419376623626, 5156596810353639551, 12813724723179037911, 10288881524157229871, 708830206584151678},
}
