function explorerChange(e) {
	if (e.selectedOptions.length < 1) {
		return;
	}

	var opt = e.selectedOptions[0];
	if (opt.dataset["error"]) {
		return;
	}

	d3.request('/chart/' + opt.value)
		.header('Content-Type', 'application/json')
		.mimeType('application/json')
		.on('error', function (xhr) {
			console.warn('couldn\'t retreive chart data', xhr);
		})
		.get(function (xhr) {
			updateChart(JSON.parse(xhr.response));
		});
}

function updateChart(data) {
	var measures = data["Measures"];
	var interval = ((Date.parse(measures["End"]) - Date.parse(measures["Start"])) / 1000) / measures.Data.length;
	liveChart.init("#chart", [measures.Data,data["Measures1"].Data,data["Measures2"].Data,data["Measures3"].Data], false, interval);

	d3.selectAll('.cyType').html(data.CycleType);
	d3.selectAll('.cyStatus').html(data.Reason);
	d3.selectAll('.cyRuntime').html(data.TotalDuration);
	d3.selectAll('.cyInterval').html(measures.Interval);
	d3.selectAll('.cyStart').html(measures.Start);
	d3.selectAll('.cyEnd').html(measures.End);

	var cfg = data["Config"];
	d3.selectAll('.cfgMode').html(cfg.Mode);
	d3.selectAll('.cfgNbHalfCycles').html(cfg.NbHalfCycles);
	d3.selectAll('.cfgUpDuration').html(cfg.UpDuration);
	d3.selectAll('.cfgDownDuration').html(cfg.DownDuration);
	d3.selectAll('.cfgTopVoltage').html(cfg.TopVoltage);
	d3.selectAll('.cfgBottomVoltage').html(cfg.BottomVoltage);
	d3.selectAll('.cfgTicker').html(cfg.Ticker);
	d3.selectAll('.cfgChargeFirst').html(cfg.ChargeFirst);

	var user = data["User"];
	d3.selectAll(".userId").html(user.BetaId);
	d3.selectAll(".userName").html(user.Name);
	var battery = data["Battery"];
	d3.selectAll(".batteryRef").html(battery.BetaRef);
	d3.selectAll(".batteryType").html(battery.Type);
	if (battery.Voltage) {
		d3.selectAll(".batteryVoltage").html("" + battery.Voltage + "mV");
	} else {
		d3.selectAll(".batteryVoltage").html("");
	}

	d3.selectAll(".batteryBrand").html(battery.Brand);
	d3.selectAll(".batteryModel").html(battery.Model);
	var res = data["Resistor"];
	if (res) {
		d3.selectAll(".resValue").html("" + res + "ohm");
	} else {
		d3.selectAll(".resValue").html("");
	}
}
