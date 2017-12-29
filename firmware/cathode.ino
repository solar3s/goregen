/*---------------------------------------------------------------------*
  This file comes pre-installed in your Regenbox, it is responsible
  for communicating with the main executable "goregen"

  In case of a major upgrade, or if you need to install it on a new
  Arduino board, please refer to the wiki section of the project :

      https://github.com/solar3s/goregen/wiki/Upgrading-firmware
-----------------------------------------------------------------------*/

#define VERSION "cathode"

/*---------------------------------------------------------------------*
  Provides direct pin access via simple serial protocol

  Takes a 1byte instruction input, output can be:
    - On toggle instructions: write 1 single boolean byte for new state
        (LED_TOGGLE)
    - On other pin instructions: write 1 single 0 byte (always success)
        (LED_0/1, PIN_DISCHARGE_0/1, PIN_CHARGE_0/1, MODE_*)
    - On uint readings: write ascii repr of value
        (READ_A0, READ_V)

  All communication end with STOP_BYTE

-----------------------------------------------------------------------*/


// -------------------------------------------
// Input instructions (waiting on serial read)
//
// READ_* writes string response
#define READ_A0         0x00 // read A0 pin
#define READ_V          0x01 // fancy A0 reads and compute voltage
#define READ_VERSION    0x02 // returns current firmware version

// LED_TOGGLE writes boolean response (led state)
#define LED_TOGGLE      0x12 // led toggle

// all other commands return a single null byte
#define LED_0           0x10 // led off
#define LED_1           0x11 // led on
#define PIN_DISCHARGE_0 0x30 // pin discharge off
#define PIN_DISCHARGE_1 0x31 // pin discharge on
#define PIN_CHARGE_0    0x40 // pin charge off
#define PIN_CHARGE_1    0x41 // pin charge on

#define MODE_IDLE       0x50 // enable idle mode
#define MODE_CHARGE     0x51 // enable charge mode
#define MODE_DISCHARGE  0x52 // enable discharge mode

#define PING            0xA0 // just a ping

#define STOP_BYTE       0xff // sent after all communication

// ---------------------------
// Internal address and config
#define PIN_CHARGE    4        // output pin address (charge)
#define PIN_DISCHARGE 3        // output pin address (discharge)
#define PIN_LED       13       // output pin address (arduino led)
#define PIN_ANALOG    A0       // analog pin on battery-0 voltage

// config parameters for getVoltage()
#define CAN_REF       2410 // tension de reference du CAN
unsigned long AREF;
#define CAN_BITSIZE   1023 // précision du CAN
#define NB_ANALOG_RD  204  // how many analog read to measure average

// Averaging parameters
#define VOLTAGE_HISTORY_NUM  10                  // Number of samples for averaging
unsigned long gVoltageHist[VOLTAGE_HISTORY_NUM]; // Voltage history
unsigned long gHistCounter = 0;                  // Voltage measurement counter

// computeAvgVoltage retreive the previous last
// VOLTAGE_HISTORY_NUM measures and averages on that
unsigned long computeAvgVoltage() {
  unsigned long avgVoltage = 0;
  byte sz = gHistCounter < VOLTAGE_HISTORY_NUM?
    gHistCounter: VOLTAGE_HISTORY_NUM;
  for (byte i = 0; i < sz; i++) {
    avgVoltage += gVoltageHist[i];
  }
  avgVoltage = floor(avgVoltage / sz);
  return avgVoltage;
}

void setCharge(boolean b) {
  digitalWrite(PIN_CHARGE, !b);
}

void setDischarge(boolean b) {
  digitalWrite(PIN_DISCHARGE, b);
}

void setLed(boolean b) {
  digitalWrite(PIN_LED, b);
}

boolean toggleLed() {
  boolean b = !digitalRead(PIN_LED);
  setLed(b);
  return b;
}

unsigned long getAnalog() {
  return analogRead(PIN_ANALOG);
}

unsigned long getVoltage() {
  unsigned long tmp, sum;
  for(byte i=0; i < NB_ANALOG_RD; i++){
    tmp = getAnalog();
    sum = sum + tmp;
    delay(1);
  }
  sum = sum / NB_ANALOG_RD;
  // convert using CAN specs and ref value
  sum = (sum * AREF) / CAN_BITSIZE;

  gVoltageHist[gHistCounter % VOLTAGE_HISTORY_NUM] = sum;
  gHistCounter++;
  
  return computeAvgVoltage();
}

// initAref is a guru trick taken on : https://forum.arduino.cc/index.php?topic=267827.msg1889127#msg1889127
// it retreives AREF value (otherwise unavailable for reading) by doing registry & mux manipulation tricks.
unsigned long initAref() {
  float volt;

#if defined (__AVR_ATmega8__)
#elif defined (__AVR_ATmega168__)
#elif defined (__AVR_ATmega168A__)
#elif defined (__AVR_ATmega168P__)
#elif defined (__AVR_ATmega328__)
#elif defined (__AVR_ATmega328P__)

  // set reference to AREF, and mux to read the internal 1.1V
  // REFS1 = 0, REFS0 = 0, MUX3..0 = 1110
  ADMUX = _BV(MUX3) | _BV(MUX2) | _BV(MUX1);

  // Enable the ADC
  ADCSRA |= _BV(ADEN);

  // Wait for voltage to become stable after changing the mux.
  delay(20);

  // Start ADC
  ADCSRA |= _BV(ADSC);

  // wait for the ADC to finish
  while (bit_is_set(ADCSRA, ADSC));

  // Read the ADC result
  // The 16-bit ADC register is 'ADC' or 'ADCW'
  unsigned int raw = ADCW;

  // Calculate the Aref.
  volt = 1100.0 / (float) raw * 1024.0;

#elif defined (__AVR_ATmega32U4__)
#elif defined (__AVR_ATmega1280__) || defined (__AVR_ATmega2560__)
#endif

  // Try to return to normal.
  analogReference(EXTERNAL);
  analogRead(A0);            // the mux is set, throw away ADC value
  delay(20);                 // wait for voltages to become stable

  return floor(volt);
}

void setup() {
  Serial.begin(57600);
  // reference de tension pour les mesures
  AREF = initAref();

  pinMode(PIN_CHARGE, OUTPUT);
  pinMode(PIN_DISCHARGE, OUTPUT);
  pinMode(PIN_LED, OUTPUT);

  setCharge(0);
  setDischarge(0);
  setLed(1);
}


// simple talk protocol
void loop() {
  if (!Serial.available()) {
    return;
  }

  byte in = Serial.read();
  switch (in) {
    case READ_VERSION:
      Serial.print(VERSION);
      break;
    case READ_A0:
      Serial.print(getAnalog());
      break;
    case READ_V:
      Serial.print(getVoltage());
      break;

    case LED_0:
      setLed(0);
      break;
    case LED_1:
      setLed(1);
      break;
    case LED_TOGGLE:
      Serial.write(toggleLed());
      break;

    case PIN_DISCHARGE_0:
      setDischarge(0);
      break;
    case PIN_DISCHARGE_1:
      setDischarge(1);
      break;

    case PIN_CHARGE_0:
      setCharge(0);
      break;
    case PIN_CHARGE_1:
      setCharge(1);
      break;

    case MODE_IDLE:
      setDischarge(0);
      setCharge(0);
      break;
    case MODE_CHARGE:
      setDischarge(0);
      setCharge(1);
      break;
    case MODE_DISCHARGE:
      setCharge(0);
      setDischarge(1);
      break;
    case PING:
      break;
    default:
      // do not talk to strangers
      return;
  }

  // end communication
  Serial.write(STOP_BYTE);
}
