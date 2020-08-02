function grayMultiVolatages(){
	d3.selectAll('.vVoltage1').style('color', 'gray');
	d3.selectAll('.vVoltage2').style('color', 'gray');
	d3.selectAll('.vVoltage3').style('color', 'gray');
}

function rbCharge() {
	grayMultiVolatages();
	setConfig({
		Mode: "Charger"
	}, rbStart);
}

function rbChargeX4() {
	// checks
	let one=false;
	if(d3.selectAll('.cBattery1').node().checked)one=true;
	else if(d3.selectAll('.cBattery2').node().checked)one=true;
	else if(d3.selectAll('.cBattery3').node().checked)one=true;
	else if(d3.selectAll('.cBattery4').node().checked)one=true;
	if(!one){
		alert("Please select at least one battery");
		return;
	}
	if(!d3.selectAll('.cBattery1').node().checked)d3.selectAll('.vVoltage').style('color', 'gray');
	if(!d3.selectAll('.cBattery2').node().checked)d3.selectAll('.vVoltage1').style('color', 'gray');
	if(!d3.selectAll('.cBattery3').node().checked)d3.selectAll('.vVoltage2').style('color', 'gray');
	if(!d3.selectAll('.cBattery4').node().checked)d3.selectAll('.vVoltage3').style('color', 'gray');
	setConfig({
		Mode: "ChargerX4",
		Battery1: d3.selectAll('.cBattery1').node().checked,
		Battery2: d3.selectAll('.cBattery2').node().checked,
		Battery3: d3.selectAll('.cBattery3').node().checked,
		Battery4: d3.selectAll('.cBattery4').node().checked
	}, rbStart);
}

function rbDischarge() {
	grayMultiVolatages();
	setConfig({
		Mode: "Discharger"
	}, rbStart);
}

function rbCycle() {
	grayMultiVolatages();
	setConfig({
		Mode: "Cycler"
	}, rbStart);
}

function setConfig(cfg, callback) {
	if (typeof(cfg) === 'object') {
		cfg = JSON.stringify(cfg);
	}
	if (typeof(cfg) !== 'string') {
		throw new Error('expecting cfg to be a string, got ', typeof(cfg));
	}

	d3.request('/config?save')
		.header('Content-Type', 'application/json')
		.mimeType('application/json')
		.on('error', function(xhr) {
			console.warn('error in setConfig', xhr);
		})
		.post(cfg, function(xhr) {
			cfg = JSON.parse(xhr.response);
			d3.selectAll('.cfgMode').html(cfg.Mode);
			d3.selectAll('.cfgNbHalfCycles').html(cfg.NbHalfCycles);
			d3.selectAll('.cfgUpDuration').html(cfg.UpDuration);
			d3.selectAll('.cfgDownDuration').html(cfg.DownDuration);
			d3.selectAll('.cfgTopVoltage').html(cfg.TopVoltage);
			d3.selectAll('.cfgBottomVoltage').html(cfg.BottomVoltage);
			d3.selectAll('.cfgTicker').html(cfg.Ticker);
			d3.selectAll('.cfgChargeFirst').html(cfg.ChargeFirst);
			callback(cfg);
		});
}

function rbStop() {
	d3.request('/stop')
		.on('error', function (xhr) {
			console.warn('error in stop', xhr);
		})
		.post({}, function() {
			d3.selectAll('.ctrl.cUp').attr('disabled', null);
			d3.selectAll('.ctrl.cDown').attr('disabled', '');
		});
}

function rbStart() {
	d3.request('/start')
		.on('error', function (xhr) {
			console.warn('error in start', xhr);
		})
		.post({}, function() {
			d3.selectAll('.ctrl.cUp').attr('disabled', '');
			d3.selectAll('.ctrl.cDown').attr('disabled', null);
		});
}