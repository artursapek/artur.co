require 'date'
require 'optparse'

album_file = ARGV[0]
caption_style = "before"


OptionParser.new do |opts|
  opts.on("-c", "--captions", "Caption style") do |v|
    if v != "before" && v != "after"
      puts "fatal: --captions option must be either \"before\" or \"after\""
      exit 1
    end
    caption_style = v
  end
end.parse!

title = nil
date = nil
images = []

pending_caption = nil

puts "Parsing album at #{album_file}"

File.open(album_file, "r") do |f|
  f.each_line do |line|
    if line =~ /h1/
      title = line.gsub(/<\/?h1>/,"").split.map(&:capitalize).join(' ')
    elsif line =~ /h3/
      date = Date.parse(line.gsub(/<\/?h3>/,"").split.map(&:capitalize).join(' '))
    elsif line =~ /<p>/
      caption = line.gsub(/<\/?p>/,"").gsub("\n","")

      if caption_style == "before"
        pending_caption = caption
      else
        images[-1][:caption] = caption
      end

    elsif line =~ /img.*src/
      src = /src="([^"]*)"/.match(line)[1]
      src.gsub!("./","")
      fn = src.split(".")[0]
      image = { fn: fn }
      if pending_caption
        image[:caption] = pending_caption
        pending_caption = nil
      end
      images << image

    end
  end
end

yml_file = [
  "title: #{title}",
  "date: #{date.strftime "%-m/%-d/%y"}",
  "content:"
]

puts images

puts yml_file

images[0...5].each do |img|
  fn = img[:fn]
  puts fn
  paths = `sshb artur "cd /mnt/raw/photos; find . -name *#{fn}*"`.split("\n")
  # For the purposes of this script, we can omit anything
  # before 2014 to eliminate duplicate filenames
  paths = paths.select{ |p| p.split("/")[1].to_i > 2013 }

  if paths.size > 1
    puts "Ambiguous fn #{fn}; more than one found path"
    puts paths.map{|p| "- #{p}"}
    puts ""
  else
    path = paths[0]
    year = path.split("/")[1]
    month = path.split("/")[2]
    yml_file << "- type: photo"
    yml_file << "  src: #{[year, month, "#{fn}.JPG"].join("/")}"
    if img[:caption]
      yml_file << "  caption: \"#{img[:caption]}\""
    end
  end
end

if title.nil?
  raise "Couldn't find title h1 element!"
elsif date.nil?
  raise "Couldn't find date h3 element!"
end

slug = title.split(" ").map(&:downcase).join("-")

f = File.new("/usr/local/go/src/github.com/artursapek/artur.co/content/photos/albums/#{slug}.yml", "w")
f.write(yml_file.join("\n"))
f.close
