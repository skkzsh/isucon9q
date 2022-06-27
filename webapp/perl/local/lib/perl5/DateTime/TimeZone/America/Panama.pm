# This file is auto-generated by the Perl DateTime Suite time zone
# code generator (0.08) This code generator comes with the
# DateTime::TimeZone module distribution in the tools/ directory

#
# Generated from /tmp/tRZSIOcmOW/northamerica.  Olson data version 2019b
#
# Do not edit this file directly.
#
package DateTime::TimeZone::America::Panama;

use strict;
use warnings;
use namespace::autoclean;

our $VERSION = '2.36';

use Class::Singleton 1.03;
use DateTime::TimeZone;
use DateTime::TimeZone::OlsonDB;

@DateTime::TimeZone::America::Panama::ISA = ( 'Class::Singleton', 'DateTime::TimeZone' );

my $spans =
[
    [
DateTime::TimeZone::NEG_INFINITY, #    utc_start
59611180688, #      utc_end 1890-01-01 05:18:08 (Wed)
DateTime::TimeZone::NEG_INFINITY, #  local_start
59611161600, #    local_end 1890-01-01 00:00:00 (Wed)
-19088,
0,
'LMT',
    ],
    [
59611180688, #    utc_start 1890-01-01 05:18:08 (Wed)
60188764776, #      utc_end 1908-04-22 05:19:36 (Wed)
59611161512, #  local_start 1889-12-31 23:58:32 (Tue)
60188745600, #    local_end 1908-04-22 00:00:00 (Wed)
-19176,
0,
'CMT',
    ],
    [
60188764776, #    utc_start 1908-04-22 05:19:36 (Wed)
DateTime::TimeZone::INFINITY, #      utc_end
60188746776, #  local_start 1908-04-22 00:19:36 (Wed)
DateTime::TimeZone::INFINITY, #    local_end
-18000,
0,
'EST',
    ],
];

sub olson_version {'2019b'}

sub has_dst_changes {0}

sub _max_year {2029}

sub _new_instance {
    return shift->_init( @_, spans => $spans );
}



1;

