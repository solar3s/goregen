var scaleBottom = 800;
var scaleTop = 1600;


/*
 * data := [V][4]
 */
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
	var make_x_gridlines = function (){
		return d3.axisBottom(x)
        .ticks(5)
	}
	var make_y_gridlines = function (){
		return d3.axisLeft(y)
        .ticks(5)
	}
		
	var line = d3.line()
		.x(function (d, i) {
			return x(i);
		})
		.y(function (d, i) {
			return y(d[0]);
		});
	var line1 = d3.line()
		.x(function (d, i) {
			return x(i);
		})
		.y(function (d, i) {
			return y(d[1]);
		});
	var line2 = d3.line()
		.x(function (d, i) {
			return x(i);
		})
		.y(function (d, i) {
			return y(d[2]);
		});
	var line3 = d3.line()
		.x(function (d, i) {
			return x(i);
		})
		.y(function (d, i) {
			return y(d[3]);
		});
	var g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

	// rect
	g.append("defs").append("clipPath")
		.attr("id", "clip")
		.append("rect")
		.attr("width", width)
		.attr("height", height);
	
	// add the X gridlines
	g.append("g")			
		.attr("class", "grid")
		.attr("transform", "translate(0," + height + ")")
		.call(make_x_gridlines()
			.tickSize(-height)
			.tickFormat("")
		);

	// add the Y gridlines
	g.append("g")			
		.attr("class", "grid")
		.call(make_y_gridlines()
			.tickSize(-width)
			.tickFormat("")
		);
	
	// y axis
	g.append("g")
		.attr("class", "axis axis--y")
		.call(d3.axisLeft(y));

	// Draw data line	
	g.append("g").attr("clip-path", "url(#clip)")
		.append("path")
		.datum(data)
		.style("stroke", "black")
		.style("fill", "none")
		.attr("class", "line");
	// Draw the line.
	d3.select("path.line")
		.attr("d", line);
	
	// Draw data line1
	g.append("g").attr("clip-path", "url(#clip)")
		.append("path")
		.datum(data)
		.style("stroke", "red")
		.style("fill", "none")
		.attr("class", "line1");
	// Draw the line.
	d3.select("path.line1")
		.attr("d", line1);
	
	// Draw data line1
	g.append("g").attr("clip-path", "url(#clip)")
		.append("path")
		.datum(data)
		.style("stroke", "green")
		.style("fill", "none")
		.attr("class", "line2");
	// Draw the line.
	d3.select("path.line2")
		.attr("d", line2);
		
	// Draw data line1
	g.append("g").attr("clip-path", "url(#clip)")
		.append("path")
		.datum(data)
		.style("stroke", "blue")
		.style("fill", "none")
		.attr("class", "line3");
	// Draw the line.
	d3.select("path.line3")
		.attr("d", line3);
		
	return {
		g: g,
		data: data,
		width: width,
		height: height,
		margin: margin,
		x: x,
		y: y,
		line: line,
		line1: line1,
		line2: line2,
		line3: line3
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
			liveChart.init(selector, [data.LiveData1,data.LiveData2,data.LiveData3,data.LiveData4], true, 15);
		});
};

/*
 * data := [4][voltages]
 */
liveChart.init = function (selector, data, reverse, intervalSec) {
	if (this.svg) {
		// clear first
		d3.select(selector).html("");
	}
	// transform data [4][V] to [V][4]
	var vdata = [];
	if(!data || data.length==0)return;
	for(let i=0;i<data[0].length;i++){
		vdata.push([data[0][i],data[1][i],data[2][i],data[3][i]]);
	}
	this.data = vdata;
	this.svg = makeG(selector, vdata);
	this.tick = function (v) {
		if (!v) {
			v = Number(d3.select('.vVoltage').html().replace("mV",""));
		}
		if (!v) {
			return;
		}

		this.data.push([v,Number(d3.select('.vVoltage1').html().replace("mV","")),Number(d3.select('.vVoltage2').html().replace("mV","")),Number(d3.select('.vVoltage3').html().replace("mV",""))]);
		// Redraw the line.
		d3.select("path.line")
			.attr("d", this.svg.line)
			.attr("transform", null)
			.attr("transform", "translate(" + this.svg.x(-1) + ",0)");
		d3.select("path.line1")
			.attr("d", this.svg.line1)
			.attr("transform", null)
			.attr("transform", "translate(" + this.svg.x(-1) + ",0)");
		d3.select("path.line2")
			.attr("d", this.svg.line2)
			.attr("transform", null)
			.attr("transform", "translate(" + this.svg.x(-1) + ",0)");
		d3.select("path.line3")
			.attr("d", this.svg.line3)
			.attr("transform", null)
			.attr("transform", "translate(" + this.svg.x(-1) + ",0)");

		// Pop the old data point off the front.
		this.data.shift();
	};

	var reverseAxis = function (d) {
		if ((vdata.length - d) === 0) {
			return 'now';
		}
		return ((vdata.length - d) * intervalSec / 3600).toFixed(1) + 'h ago';
	};

	var normalAxis = function(d) {
		return moment.duration(d*intervalSec, "seconds").format("h[h]m[m]s[s]");
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
