function rbCharge() {
	setConfig({
		Mode: "Charger"
	}, rbStart);
}

function rbDischarge() {
	setConfig({
		Mode: "Discharger"
	}, rbStart);
}

function rbCycle() {
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
			d3.selectAll('.cfgIntervalSec').html(cfg.IntervalSec);
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