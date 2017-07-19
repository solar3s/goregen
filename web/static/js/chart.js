var chartMaxPoints = 1000;
var scaleBottom = 0;
var scaleTop = 1600;

function makeG(selector, data) {
	var svg = d3.select(selector);
	var margin = {top: 20, right: 20, bottom: 50, left: 40};
	var width = 0 + svg.attr("width") - margin.left - margin.right;
	var height = 0 + svg.attr("height") - margin.top - margin.bottom;
	var n = data.length;

	var x = d3.scaleLinear()
		.domain([0, n])
		.range([0, width]);
	var y = d3.scaleLinear()
		.domain([scaleBottom, scaleTop])
		.range([height, 0]);
	var line = d3.line()
		.x(function (d, i) {
			return x(i);
		})
		.y(function (d, i) {
			return y(d);
		});
	var g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

	g.append("defs").append("clipPath")
		.attr("id", "clip")
		.append("rect")
		.attr("width", width)
		.attr("height", height);

	g.append("g")
		.attr("class", "axis axis--y")
		.call(d3.axisLeft(y));

	g.append("g")
		.attr("clip-path", "url(#clip)")
		.append("path")
		.datum(data)
		.attr("class", "line");

	// Draw the line.
	d3.select("path.line")
		.attr("d", line);

	return {
		g: g,
		data: data,
		width: width,
		height: height,
		margin: margin,
		x: x,
		y: y,
		line: line
	}
}

var chart = {};
chart.init = function (data, intervalSec) {
	var n = data.length;
	if (n > chartMaxPoints) {
		// reduce number of points by ratio
		var split = n / chartMaxPoints;
		// reduce number of points to maxPoints
		n = chartMaxPoints;
		// update interval between 2 points according to ratio
		intervalSec *= split;
		// only take 1 point from data every split
		var data2 = [];
		for (var i = 0; i < chartMaxPoints; i++) {
			data2.push(data[(i*split).toFixed(0)]);
		}
		data = data2;
	}

	this.svg = makeG("#chart", data);
	return this;
};

var liveChart = {};
liveChart.init = function (zeroValue, n) {
	if (!zeroValue) {
		zeroValue = 0;
	}
	if (!n) {
		n = 2400;
	}

	var data = [];
	for (var i = n; i > 0; i--) data.push(zeroValue);

	this.svg = makeG("#chart", data);
	this.tick = function () {
		var v = Number(d3.select('.vRawVoltage').html());
		if (!v) {
			return;
		}

		data.push(v);
		// Redraw the line.
		d3.select("path.line")
			.attr("d", this.svg.line)
			.attr("transform", null)
			.attr("transform", "translate(" + this.svg.x(-1) + ",0)");

		// Pop the old data point off the front.
		data.shift();
	};

	this.svg.g.append("g")
		.attr("class", "axis axis--x")
		.attr("transform", "translate(0," + this.svg.y(scaleBottom) + ")")
		.call(d3.axisBottom(this.svg.x).tickFormat(function (d) {
			if ((n - d) === 0) {
				return 'now';
			}
			return ((n - d) / (60 * 4)).toFixed(1) + 'h ago';
		}))
		.selectAll("text")
		.style("text-anchor", "end")
		.attr("dx", "-.8em")
		.attr("dy", ".15em")
		.attr("transform", "rotate(-45)");
};
