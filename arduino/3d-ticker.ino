#include <FastLED.h>

// How many leds in your strip?
#define NUM_LEDS_PER_STRIP 10
#define NUM_STRIPS 5
#define BRIGHTNESS 96

// available leds
CRGB leds[NUM_STRIPS][NUM_LEDS_PER_STRIP];

// values to display in the bars
int bar_values[NUM_STRIPS];

int daily_high;

void setup() {
  Serial.begin(9600);

  FastLED.addLeds<WS2811, 2, GRB>(leds[0], NUM_LEDS_PER_STRIP);
  FastLED.addLeds<WS2811, 3, GRB>(leds[1], NUM_LEDS_PER_STRIP);
  FastLED.addLeds<WS2811, 4, GRB>(leds[2], NUM_LEDS_PER_STRIP);
  FastLED.addLeds<WS2811, 5, GRB>(leds[3], NUM_LEDS_PER_STRIP);
  FastLED.addLeds<WS2811, 6, GRB>(leds[4], NUM_LEDS_PER_STRIP);
  FastLED.setBrightness(BRIGHTNESS);

  //init_daily_values_rand();

}

void loop() {

  update_daily_values_serial();
  update_daily_high();

  for (int strip = 0; strip <= NUM_STRIPS - 1; strip++) {
    for (int led = 0; led <= NUM_LEDS_PER_STRIP - 1; led++) {
      if (bar_values[strip] < led) {
        if (led == daily_high) {
          leds[strip][led] = CRGB::White;
        } else {
          leds[strip][led] = CRGB::Black;
        }
      } else {
        if (led == daily_high) {
          leds[strip][led] = CRGB::White;
        } else if (led < daily_high) {
          leds[strip][led] = CRGB::OrangeRed;
        } else {
          leds[strip][led] = CRGB::Green;
        }
      }
    }
  }
  FastLED.show();
  delay(1000);
}

void update_daily_high() {
  daily_high = bar_values[0];
}

void init_daily_values_rand() {

  daily_high = random(0, 9);

  bar_values[0] = random(0, NUM_LEDS_PER_STRIP);
  for (int strip = 1; strip <= NUM_STRIPS - 1; strip++) {
    if (bar_values[0] < 4) {
      bar_values[strip] = bar_values[0] + strip;
    } else {
      bar_values[strip] = bar_values[0] - strip;
    }
  }
}

void update_daily_values_rand() {
  for (int strip = 0; strip <= NUM_STRIPS - 1; strip++) {
    if (strip == NUM_STRIPS - 1) {
      bar_values[strip] = random(0, NUM_LEDS_PER_STRIP);
    } else {
      bar_values[strip] = bar_values[strip + 1];
    }
  }
}

void update_daily_values_serial() {
  
  char end_marker = '\n';  
  int received_vals[NUM_STRIPS];

  char rc;
  static byte ndx = 0;
  while (Serial.available() > 0) {
    rc = Serial.read();
    if (rc != end_marker) {
      // handle numbers
      if(rc >= '0' && rc <= '9') {
          Serial.println(rc);
          received_vals[ndx] = rc - '0';
          ndx++;
      }
      // handle overflow
      if (ndx >= NUM_STRIPS) {
        ndx = NUM_STRIPS - 1;
      }
    } else {
      for (int i = 0; i <= ndx; i++) {
        bar_values[i] = received_vals[i];
      }
      ndx = 0;
      return;
    }
  }
}
