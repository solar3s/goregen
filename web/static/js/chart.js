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

	// rect
	g.append("defs").append("clipPath")
		.attr("id", "clip")
		.append("rect")
		.attr("width", width)
		.attr("height", height);

	// y axis
	g.append("g")
		.attr("class", "axis axis--y")
		.call(d3.axisLeft(y));

	// Draw data line
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

var liveChart = {};
liveChart.initFrom = function(url, selector) {
	d3.request(url)
		.header('Content-Type', 'application/json')
		.mimeType('application/json')
		.on('error', function (xhr) {
			console.warn('couldn\'t retreive chart data', xhr);
		})
		.get(function (xhr) {
			var data = JSON.parse(xhr.response);
			liveChart.init(selector, data, true, 15);
		});
};

liveChart.init = function (selector, data, reverse, intervalSec) {
	if (this.svg) {
		// clear first
		d3.select(selector).html("");
	}
	this.data = data;
	this.svg = makeG(selector, data);
	this.tick = function (v) {
		if (!v) {
			v = Number(d3.select('.vRawVoltage').html());
		}
		if (!v) {
			return;
		}

		this.data.push(v);
		// Redraw the line.
		d3.select("path.line")
			.attr("d", this.svg.line)
			.attr("transform", null)
			.attr("transform", "translate(" + this.svg.x(-1) + ",0)");

		// Pop the old data point off the front.
		this.data.shift();
	};

	var reverseAxis = function (d) {
		if ((data.length - d) === 0) {
			return 'now';
		}
		return ((data.length - d) * intervalSec / 3600).toFixed(1) + 'h ago';
	};

	var normalAxis = function(d) {
		return durationString(d*intervalSec);
	};

	// x axis
	this.svg.g.append("g")
		.attr("class", "axis axis--x")
		.attr("transform", "translate(0," + this.svg.y(scaleBottom) + ")")
		.call(d3.axisBottom(this.svg.x).tickFormat(reverse ? reverseAxis : normalAxis))
		.selectAll("text")
		.style("text-anchor", "end")
		.attr("dx", "-.8em")
		.attr("dy", ".15em")
		.attr("transform", "rotate(-45)");
};

function durationString(seconds) {
	var durationString = "";
	var secs = Number(seconds).toFixed(0);
	var hours, minutes;
	if (secs >= 3600) {
		hours = Number(secs / 3600).toFixed(0);
		secs -= hours * 3600;
		durationString += "" + hours + "h ";
		minutes = Number(secs / 60).toFixed(0);
		secs -= minutes;
		durationString += "" + minutes + "m ";
	} else if (secs >= 60) {
		minutes = Number(secs / 60).toFixed(0);
		secs -= (minutes*60);
		durationString += "" + minutes + "m ";
	}
	durationString += Number(secs).toFixed(0) + "s";
	return durationString;
}
