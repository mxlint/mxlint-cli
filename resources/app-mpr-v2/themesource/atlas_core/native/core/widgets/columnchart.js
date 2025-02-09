var _a, _b, _c, _d;
import { columnChart } from "../../variables";
/*

DISCLAIMER:
Do not change this file because it is core styling.
Customizing core files will make updating Atlas much more difficult in the future.
To customize any core styling, copy the part you want to customize to styles/native/app/ so the core styling is overwritten.

==========================================================================
    Column Chart

    Default Class For Mendix Column Chart Widget
========================================================================== */
export const com_mendix_widget_native_columnchart_ColumnChart = {
    container: {
        // All ViewStyle properties are allowed
        ...columnChart.container
    },
    errorMessage: {
        // All TextStyle properties are allowed
        ...columnChart.errorMessage
    },
    chart: {
        // All ViewStyle properties are allowed
        ...columnChart.chart
    },
    grid: {
        /*
            Allowed properties:
              -  backgroundColor (string)
              -  dashArray (string)
              -  lineColor (string)
              -  padding (number)
              -  paddingBottom (number)
              -  paddingHorizontal (number)
              -  paddingLeft (number)
              -  paddingRight (number)
              -  paddingTop (number)
              -  paddingVertical (number)
              -  width (number)
        */
        ...columnChart.grid
    },
    xAxis: {
        /*
            Allowed properties:
              -  color (string)
              -  dashArray (string)
              -  fontFamily (string)
              -  fontSize (number)
              -  fontStyle ("normal" or "italic")
              -  fontWeight ("normal" or "bold" or "100" or "200" or "300" or "400" or "500" or "600" or "700" or "800" or "900")
              -  lineColor (string)
              -  width (number)
              -  label: All TextStyle properties are allowed and:
                    -relativePositionGrid ("bottom" or "right")
        */
        ...columnChart.xAxis
    },
    yAxis: {
        /*
            Allowed properties:
              -  color (string)
              -  dashArray (string)
              -  fontFamily (string)
              -  fontSize (number)
              -  fontStyle ("normal" or "italic")
              -  fontWeight ("normal" or "bold" or "100" or "200" or "300" or "400" or "500" or "600" or "700" or "800" or "900")
              -  lineColor (string)
              -  width (number)
              - label: All TextStyle properties are allowed and:
                 -  relativePositionGrid ("top" or "left")
        */
        ...columnChart.yAxis
    },
    columns: {
        /*
            Allowed properties:
                -  columnColorPalette (string with array of colors separated by ';')
                -  columnsOffset (number)
                - customColumnStyles:{
                    your_static_or_dynamic_attribute_value:{
                        column:
                            Allowed properties:
                      -  ending (number)
                      -  columnColor (string)
                      -  width (number)
                        label:
                        Allowed properties:
                      -  fontFamily (string)
                      -  fontSize (number)
                      -  fontStyle ("normal" or "italic")
                      -  fontWeight ("normal" or "bold" or "100" or "200" or "300" or "400" or "500" or "600" or "700" or "800" or "900")

                    }
                }
        */
        ...columnChart.columns
    },
    legend: {
        container: {
            // All ViewStyle properties are allowed
            ...(_a = columnChart.legend) === null || _a === void 0 ? void 0 : _a.container
        },
        item: {
            // All ViewStyle properties are allowed
            ...(_b = columnChart.legend) === null || _b === void 0 ? void 0 : _b.item
        },
        indicator: {
            // All ViewStyle properties are allowed
            ...(_c = columnChart.legend) === null || _c === void 0 ? void 0 : _c.indicator
        },
        label: {
            // All TextStyle properties are allowed
            ...(_d = columnChart.legend) === null || _d === void 0 ? void 0 : _d.label
        }
    }
};
