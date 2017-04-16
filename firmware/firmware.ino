/*-----------------------------------------------------
  Provides direct pin access via simple serial protocol

  1byte input:
    0x00 read A0 pin
    0x01 fancy A0 reads and compute voltage
    0x10 led off
    0x11 led on
    0x30 pin discharge off
    0x31 pin discharge on
    0x40 pin charge off
    0x41 pin charge on

    0x50 idle mode
    0x51 charge mode
    0x52 discharge mode

  output: depends

-----------------------------------------------------*/

#define BOX_READY     0xFF     // send when box is ready

#define PIN_CHARGE    4        // output pin (charge)
#define PIN_DISCHARGE 3        // output pin (discharge)
#define PIN_LED       13       // output pin (arduino led)
#define PIN_ANALOG    A0       // analog pin on battery-0 voltage

// config parameters for getVoltage()
#define CAN_REF       2460     // tension de reference du CAN
#define CAN_BITSIZE   1023     // précision du CAN
#define NB_ANALOG_RD  10       // how many analog read to measure average on

// some random errors
#define ERR           100

#define INTERVAL      3600000  //nombre de seconde dans une heure

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
  if (Serial.println(v) <= 0) {
    return false;
  }
  return true;
}

// sendReady sends special ready byte when setup() is over
void sendReady() {
  Serial.write(BOX_READY);
}

// boolean response
boolean sendBool(boolean v) {
  if (Serial.write(v) <= 0) {
    return false;
  }
  return true;
}

void setup() {
  Serial.begin(9600);

  pinMode(PIN_CHARGE, OUTPUT);
  pinMode(PIN_DISCHARGE, OUTPUT);
  pinMode(PIN_LED, OUTPUT);

  setCharge(0);
  setDischarge(0);
  setLed(1);

  // notify we're good to go on serial
  sendReady();
}


// simple talk protocol
//   - input: 1 byte for instruction
//   - output: string ending with CRLF (10, 13)
void loop() {
  if (!Serial.available()) {
    return;
  }

  byte in = Serial.read();
  switch (in) {
    case 0x00:  // read A0 pin
      sendUint(getAnalog()); break;
    case 0x01:  // fancy A0 reads and compute voltage
      sendUint(getVoltage()); break;

    case 0x10:  // led off
      setLed(0); sendOk(); break;
    case 0x11:  // led on
      setLed(1); sendOk(); break;
    case 0x12:  // led toggle
      sendBool(toggleLed()); break;

    case 0x30:  // pin discharge off
      setDischarge(0); sendOk(); break;
    case 0x31:  // pin discharge on
      setDischarge(1); sendOk(); break;

    case 0x40:  // pin charge off
      setCharge(0); sendOk(); break;
    case 0x41:  // pin charge on
      setCharge(1); sendOk(); break;

    case 0x50: // idle mode
      setDischarge(0);
      setCharge(0);
      sendOk();
      break;
    case 0x51: // charge mode
      setDischarge(0);
      setCharge(1);
      sendOk();
      break;
    case 0x52: // discharge mode
      setCharge(0);
      setDischarge(1);
      sendOk();
      break;
    default:
      sendError();
      break;
  }
}
