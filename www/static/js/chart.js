var d3chart = {};
d3chart.init = function (zeroValue) {
	var n = 2400;
	var data = [];
	while (n--) data.push(zeroValue);
	n = 2400;

	var svg = d3.select("svg");
	var margin = {top: 20, right: 20, bottom: 50, left: 40};
	var width = 0 + svg.attr("width") - margin.left - margin.right;
	var height = 0 + svg.attr("height") - margin.top - margin.bottom;
	var g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

	var x = d3.scaleLinear()
		.domain([0, n])
		.range([0, width]);

	var y = d3.scaleLinear()
		.domain([800, 1500])
		.range([height, 0]);

	var line = d3.line()
		.x(function (d, i) {
			return x(i);
		})
		.y(function (d, i) {
			return y(d);
		});

	function tick() {
		var v = Number(d3.select('.vRawVoltage').html());
		var shift = 0;
		if (v) {
			data.push(v);
			shift = -1;
		}

		// Redraw the line.
		d3.select(this)
			.attr("d", line)
			.attr("transform", null);
		// Slide it to the left.
		d3.active(this)
			.attr("transform", "translate(" + x(shift) + ",0)")
			.transition()
			.on("start", tick);

		if (v) {
			// Pop the old data point off the front.
			data.shift();
		}
	}

	g.append("defs").append("clipPath")
		.attr("id", "clip")
		.append("rect")
		.attr("width", width)
		.attr("height", height);

	g.append("g")
		.attr("class", "axis axis--x")
		.attr("transform", "translate(0," + y(800) + ")")
		.call(d3.axisBottom(x).tickFormat(function (d) {
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

	g.append("g")
		.attr("class", "axis axis--y")
		.call(d3.axisLeft(y));

	g.append("g")
		.attr("clip-path", "url(#clip)")
		.append("path")
		.datum(data)
		.attr("class", "line")
		.transition()
		.duration(15000)
		.ease(d3.easeLinear)
		.on("start", tick);
};
