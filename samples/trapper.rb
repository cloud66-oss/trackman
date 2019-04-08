Signal.trap("USR1") do
    puts "USR1"
end
Signal.trap("TERM") do
    puts "TERM"
end
Signal.trap("USR2") do
    puts "USR2"
end
Signal.trap("INT") do
    puts "INT"
end

puts "Waiting now..."
sleep 1000