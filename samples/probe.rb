puts "sleeping for 3 seconds before checking"
sleep 4

puts "checking for the marker"

if File.file?("marker.txt")
    puts "file is here"
    puts "deleting marker"
    File.delete("marker.txt")
    exit(0)
else
    exit(1)
end
