require 'fileutils'

puts "sleeping for 2 seconds"
sleep 2

puts "dropping a marker"
FileUtils.touch('marker.txt')

puts "leaving"

exit(0)