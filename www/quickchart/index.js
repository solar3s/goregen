function readTextFile(file, callback)
{
  var rawFile = new XMLHttpRequest();
  rawFile.open("GET", file, false);
  rawFile.onreadystatechange = function ()
  {
    if(rawFile.readyState === 4)
    {
      if(rawFile.status === 200 || rawFile.status == 0)
      {
        var allText = rawFile.responseText;
        callback(allText);
      }
    }
  }
  rawFile.send(null);
}

function parseChartSerie(input) {
  var data = input.split("\n");
  data.pop();
  var labels = Array.from(' '.repeat(data.length));
  for (var i = 1; i < (labels.length / 60); i++) {
    labels[i*60] = ""+ i + "h";
  }

  var args = {
    labels: labels,
    series: [
      {
        data: data,
      }
    ],
  }
  return args;
}

function setFullScreen(elemId) {
  var w = window,
      d = document,
      e = d.documentElement,
      g = d.getElementsByTagName('body')[0],
      x = w.innerWidth || e.clientWidth || g.clientWidth,
      y = w.innerHeight|| e.clientHeight|| g.clientHeight;
  var target = document.getElementById(elemId);
  target.style = "width: " + 0.96*x + "px; height: " + 0.98*y + "px; margin: auto;";
}

window.onload = function() {
  setFullScreen("chart");

  // some options
  var options = {
    showPoint: false,
  };

  // some responsive options
  var responsiveOptions = [
    ['screen and (min-width: 641px) and (max-width: 1024px)', {
      showPoint: false,
      axisX: {
        labelInterpolationFnc: function(value) {
          return 'Week ' + value;
        }
      }
    }],
  ['screen and (max-width: 640px)', {
    showLine: false,
    axisX: {
      labelInterpolationFnc: function(value) {
        return 'W' + value;
      }
    }
  }]];

  // the Math.random part is a simple trick to avoid 304 & other caches
  readTextFile("/data.log?" + (0|Math.random()*9e6).toString(36), function(txtSerie) {
    var data = parseChartSerie(txtSerie);
    var chart = new Chartist.Line('#chart', data, options);
  });
}

