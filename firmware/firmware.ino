/*----------------------------------------------------------------------
  Provides direct pin access via simple serial protocol

  Takes a 1byte instruction input, output can be:
    - On toggle instructions: write 1 single boolean byte for new state
        (LED_TOGGLE)
    - On other pin instructions: write 1 single 0 byte (always success)
        (LED_0/1, PIN_DISCHARGE_0/1, PIN_CHARGE_0/1, MODE_*)
    - On uint readings: write ascii repr of value
        (READ_A0, READ_V)
-----------------------------------------------------------------------*/


// -------------------------------------------
// Input instructions (waiting on serial read)
//
// READ_* writes string response
#define READ_A0         0x00 // read A0 pin
#define READ_V          0x01 // fancy A0 reads and compute voltage

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

// default return values
#define OK              0x00
#define ERR             0x10

// ---------------------------
// Internal address and config
#define PIN_CHARGE    4        // output pin address (charge)
#define PIN_DISCHARGE 3        // output pin address (discharge)
#define PIN_LED       13       // output pin address (arduino led)
#define PIN_ANALOG    A0       // analog pin on battery-0 voltage

// config parameters for getVoltage()
#define CAN_REF       2460     // tension de reference du CAN
#define CAN_BITSIZE   1023     // précision du CAN
#define NB_ANALOG_RD  10       // how many analog read to measure average on


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
  }
  sum = sum / NB_ANALOG_RD;
  // conver using CAN specs and ref value
  sum = (sum * CAN_REF) / CAN_BITSIZE;
  return sum;
}

// standard ok response
boolean sendOk() {
  if (Serial.write(0) != 1) {
    // nothing was written. it's a shame
    return false;
  }
  return true;
}

// generic useless error
boolean sendError() {
  if (Serial.write(ERR) != 1) {
    // nothing was written. it's a shame
    return false;
  }
  return true;
}

// uint response
boolean sendUint(unsigned long v) {
  if (Serial.print(v) <= 0) {
    return false;
  }
  return true;
}

// boolean response
boolean sendBool(boolean v) {
  if (Serial.write(v) <= 0) {
    return false;
  }
  return true;
}

void setup() {
  Serial.begin(57600);

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
    case READ_A0:
      sendUint(getAnalog());
      break;
    case READ_V:
      sendUint(getVoltage());
      break;

    case LED_0:
      setLed(0);
      sendOk();
      break;
    case LED_1:
      setLed(1);
      sendOk();
      break;
    case LED_TOGGLE:
      sendBool(toggleLed());
      break;

    case PIN_DISCHARGE_0:
      setDischarge(0);
      sendOk();
      break;
    case PIN_DISCHARGE_1:
      setDischarge(1);
      sendOk();
      break;

    case PIN_CHARGE_0:
      setCharge(0);
      sendOk();
      break;
    case PIN_CHARGE_1:
      setCharge(1);
      sendOk();
      break;

    case MODE_IDLE:
      setDischarge(0);
      setCharge(0);
      sendOk();
      break;
    case MODE_CHARGE:
      setDischarge(0);
      setCharge(1);
      sendOk();
      break;
    case MODE_DISCHARGE:
      setCharge(0);
      setDischarge(1);
      sendOk();
      break;
    case PING:
      sendOk();
      break;
    default:
      sendError();
      break;
  }
}
