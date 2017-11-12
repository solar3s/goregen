var stateSocket = {};
stateSocket.init = function(addr) {
	if (addr) {
		this.listenAddr = addr;
	}
	this.reconnectButton = '<button onclick="stateSocket.init();">Reconnect</button>';
	d3.selectAll('.ctrl').attr('disabled', true);
	d3.selectAll('.ws').html('connecting...');
	var ws = new WebSocket('ws://' + this.listenAddr + '/websocket');
	var wsError = setTimeout(function () {
		d3.selectAll('.vState').html('no connection to goregen');
		var err = 'couldn\'t connect to goregen server, is it running?';
		console.warn(err);
		d3.selectAll('.ws').html(err);
		setTimeout(function () {
			d3.selectAll('.ws').html(stateSocket.reconnectButton);
		}, 5000);
	}, 1500);
	ws.onopen = function () {
		clearInterval(wsError);
		stateSocket.ws = ws;
		stateSocket.start();
	};
};

var runtime = 0;
stateSocket.start = function() {
	var ws = this.ws;
	d3.selectAll('.ws').html('Ok');
	ws.onerror = function (e) {
		console.log('websocket error', e);
	};
	ws.onmessage = function (e) {
		var v = JSON.parse(e.data);
		switch(v.Type) {
			case "ticker":
				if (liveChart.tick) {
					liveChart.tick(v.Data);
				} else {
					console.warn('liveChart not initialized');
				}
				return;
			case "state":
				var state = v.Data['State'];
				d3.selectAll('.vState').html(state);
				if (state !== 'Connected') {
					d3.selectAll('.vVoltage').html('-');
					d3.selectAll('.vRawVoltage').html('');
					d3.selectAll('.vChargeState').html('-');
					d3.selectAll('.vFirmware').html('-');
					d3.selectAll('.ctrl').attr('disabled', '');
					return;
				}

				var charge = v.Data['ChargeState'];
				d3.selectAll('.vVoltage').html(v.Data['Voltage'] + 'mV');
				d3.selectAll('.vRawVoltage').html(v.Data['Voltage']);
				d3.selectAll('.vChargeState').html(charge);
				d3.selectAll('.vFirmware').html(v.Data['Firmware']);
				d3.selectAll('.ctrl.cUp').attr('disabled', charge !== 'Idle' ? '' : null);
				d3.selectAll('.ctrl.cDown').attr('disabled', charge === 'Idle' ? '' : null);
				return;
			case "cycle":
				var cy = v.Data;
				if (cy['Final']) {
					d3.selectAll('.cycleREC').classed('hidden', true);
					clearInterval(runtime);
					runtime = 0;
				} else {
					d3.selectAll('.cycleREC').classed('hidden', false);
					if (runtime === 0) {
						var start = Date.now();
						runtime = setInterval(function () {
							var delta = Date.now() - start;
							d3.selectAll('.cyRuntime').html('' + Math.floor(delta / 1000) + ' sec');
						}, 1000);
					}
				}
				d3.selectAll('.cyType').html(cy['Type']);
				d3.selectAll('.cyStatus').html(cy['Status']);
				d3.selectAll('.cyTarget').html(cy['Target'] + 'mV');
				return;
			default:
				console.error("unknown event", v);
				return;
		}
	};
	ws.onclose = function (e) {
		d3.selectAll('.vState').html('no connection to goregen');
		d3.selectAll('.vVoltage').html('-');
		d3.selectAll('.vRawVoltage').html('');
		d3.selectAll('.vChargeState').html('-');
		d3.selectAll('.ctrl').attr('disabled', '');
		d3.selectAll('.ws').html(stateSocket.reconnectButton);
		ws.onclose = null;
		ws.onerror = null;
		ws.onmessage = null;
	}
};
