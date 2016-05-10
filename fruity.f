q1 query friend

color name color
color tag variable
color query q1

friend name friend
friend tag variable
friend query q1

fruit name fruit
fruit tag variable
fruit query q1


s1 tag scan
s1 query q1

s1b1 field entity
s1b1 variable fruit
s1b1 source s1
s1b1 tag binding

s1b2 field attribute
s1b2 constant color
s1b2 source s1
s1b2 tag binding

s1b3 field value
s1b3 variable color
s1b3 source s1
s1b3 tag binding

s2 tag scan
s2 query q1

s2b1 field entity
s2b1 variable fruit
s2b1 source s2
s2b1 tag binding

s2b2 field attribute
s2b2 constant tag
s2b2 source s2
s2b2 tag binding

s2b3 field value
s2b3 constant color
s2b3 source s2
s2b3 tag binding


s3 tag scan
s3 query q1

s3b1 field entity
s3b1 variable fruit
s3b1 source s3
s3b1 tag binding

s3b2 field attribute
s3b2 constant color
s3b2 source s3
s3b2 tag binding

s3b3 field value
s3b3 variable color
s3b3 source s3
s3b3 tag binding

s4 tag scan
s4 query q1

s4b1 field entity
s4b1 variable fruit
s4b1 source s4
s4b1 tag binding

s4b2 field attribute
s4b2 constant tag
s4b2 source s4
s4b2 tag binding

s4b3 field value
s4b3 constant color
s4b3 source s4
s4b3 tag binding


e1 expression q1
e1 operator not=

e1b1 field a
e1b1 variable friend
e1b1 source e1
e1b1 tag binding

e1b2 field b
e1b2 variable fruit
e1b2 source e1
e1b2 tag binding
