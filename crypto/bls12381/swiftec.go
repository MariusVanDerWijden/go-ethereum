package bls12381

func ecMapG1(u *fe) (*fe, *fe) {
	params := eccParamsForG1
	var t = new(fe) // TODO figure out what to do here
	var s = new(fe)

	// Evaluate initial point in conic X(u),Y(u)
	var X0 = new(fe)
	mul(X0, params.X[2], u)
	add(X0, X0, params.X[1])
	mul(X0, X0, u)
	add(X0, X0, params.X[0])
	var Y0 = new(fe)
	mul(Y0, params.Y[1], u)
	add(Y0, Y0, params.Y[0])

	// Evaluate f(u)=3u^2+4a and g(u)=u^3+au+b
	var f = new(fe)
	square(f, u)
	var g = new(fe)
	add(g, f, params.a)
	mul(g, g, u)
	add(g, g, params.b)
	var Z1 = new(fe)
	add(Z1, f, f)
	add(f, Z1, f)
	add(f, f, params.ax4)

	// Compute new point in conic with
	// X1 = f*(Y0-t*X0)^2 + g
	// Z1 = X0(1 + f*t^2)
	// Y1 = Z1*Y0 + t*(X - Z*X0)
	mul(Z1, t, X0)
	var Y1 = new(fe)
	sub(Y1, Y0, Z1)
	var X1 = new(fe)
	square(X1, Y1)
	mul(X1, X1, f)
	add(X1, X1, g)
	mul(Z1, Z1, t)
	mul(Z1, Z1, f)
	add(Z1, Z1, X0)
	mul(Y1, Y1, Z1)
	var tX = new(fe)
	mul(tX, t, X1)
	add(Y1, Y1, tX)

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
	if c2 := !isQuadraticNonResidue(y22); c2 {
		x1, x2 = x1, x2 // TODO this probably makes this non-constant time
		y21, y22 = y21, y22
	} else {
		x1, x2 = x2, x1
		y21, y22 = y22, y21
	}

	if c3 := !isQuadraticNonResidue(y23); c3 {
		x1, x3 = x1, x3 // TODO non-constant time
		y21, y23 = y21, y23
	} else {
		x1, x3 = x3, x1
		y21, y23 = y23, y21
	}

	// Find the square-root and choose sign
	sqrt(y21, y21)
	neg(y22, y21)
	if c1 := y21.sign() == s.sign(); c1 {
		y21, y22 = y21, y22
	} else {
		y21, y22 = y22, y21
	}
	return x1, y21
}

var eccParamsForG1 = struct {
	X   [3]*fe
	Y   [2]*fe
	ax4 *fe
	a   *fe
	b   *fe
}{
	[3]*fe{b, b, b},
	[2]*fe{b, b},
	b,
	b,
	b,
}
