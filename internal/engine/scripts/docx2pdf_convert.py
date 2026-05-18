import sys
from docx2pdf import convert

src, out_dir = sys.argv[1], sys.argv[2]
convert(src, out_dir)
