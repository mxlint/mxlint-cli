import { brand } from "../../variables";
/*

DISCLAIMER:
Do not change this file because it is core styling.
Customizing core files will make updating Atlas much more difficult in the future.
To customize any core styling, copy the part you want to customize to styles/native/app/ so the core styling is overwritten.

==========================================================================
    Bar Chart

    Default Class For Mendix Bar Chart Widget
========================================================================== */
export const com_mendix_widget_native_piedoughnutchart_PieDoughnutChart = {
    container: {
        // All ViewStyle properties are allowed
        flex: 1
    },
    slices: {
        /*
            Allowed properties:
                -  colorPalette (string with array of colors separated by ';')
                -  innerRadius (number)
                -  padding (number)
                -  paddingBottom (number)
                -  paddingHorizontal (number)
                -  paddingLeft (number)
                -  paddingRight (number)
                -  paddingTop (number)
                -  paddingVertical (number)
        */
        padding: 40,
        colorPalette: Object.entries(brand)
            .reduce((accumulator, [key, value]) => (key.endsWith("Light") ? accumulator : [...accumulator, value]), [])
            .join(";"),
        customStyles: {
            your_defined_key: {
                slice: {
                /*
                Allowed properties:
                  -  color (string)
            */
                },
                label: {
                /*
                Allowed properties:
                  -  color (string)
                  -  fontFamily (string)
                  -  fontSize (number)
                  -  fontStyle ("normal" or "italic")
                  -  fontWeight ("normal" or "bold" or "100" or "200" or "300" or "400" or "500" or "600" or "700" or "800" or "900")
                */
                }
            }
        }
    }
};
