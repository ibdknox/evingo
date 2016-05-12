

s1 tag scan

s1b1 tag binding
s1b1 source s1
s1b1 field entity

s1b2 tag result
s1b2 field value
s1b2 register 7

s1 next s2

s2 tag sum
s2b1 tag binding
s2b1 source s1
s2b1 field result
s2b1 register 7

s2b2 tag binding
s2b2 source s1
s2b2 field source
s2b2 register 7

s2 next s3

s3 tag plus

s3b1 tag binding
s3b1 source s3
s3b1 field result
s3b1 register 7

s3b2 tag binding
s3b2 source s3
s3b1 field a
s3b1 constant 7

s3b3 tag binding
s3b3 source s3
s3b3 field a
s3b3 register 3

